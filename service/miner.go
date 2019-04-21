package service

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/utils"
	"net"
	"sync"
)

var (
	instance  *PipeMiner = nil
	once      sync.Once
	logger, _ = logging.GetLogger(utils.LMService)
)

type PipeCmd int

const (
	MaxCharger         = 100
	_          PipeCmd = iota
	CmdPayChanel
	CmdPipe
	CmdCheck
)

type YouPipeACK struct {
	Success bool
	Message string
}

type YPHandShake struct {
	CmdType PipeCmd
	Sig     []byte
	Pipe    *PipeReqData
	Lic     *License
}

type PipeMiner struct {
	sync.RWMutex
	done       chan error
	proofSaver *Receipt
	serverConn net.Listener
	chargers   map[string]*bandCharger
}

func GetMiner() *PipeMiner {
	once.Do(func() {
		instance = newMiner()
	})

	return instance
}

func newMiner() *PipeMiner {
	acc := account.GetAccount()
	addr := network.JoinHostPort(Config.ServiceIP, acc.Address.ToServerPort())
	s, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("YouPipe miner server invalid:%s", err)
	}

	logger.Infof("Service mining at:%s", addr)

	node := &PipeMiner{
		done:       make(chan error),
		proofSaver: newReceipt(),
		serverConn: s,
		chargers:   make(map[string]*bandCharger),
	}

	return node
}

func (node *PipeMiner) Mining() {

	go node.proofSaver.DBWork()

	go node.PatrolLicense()

	defer node.serverConn.Close()

	for {
		conn, err := node.serverConn.Accept()
		if err != nil {
			panic(err)
		}
		conn.(*net.TCPConn).SetKeepAlive(true)
		c := &JsonConn{conn}
		go node.Serve(c)
	}
}

func (node *PipeMiner) Serve(conn *JsonConn) {
	hs := &YPHandShake{}
	if err := conn.ReadJsonMsg(hs); err != nil {
		conn.Close()
		return
	}

	switch hs.CmdType {
	case CmdCheck:
		node.answerCheck(conn)
	case CmdPipe:
		node.pipeServe(conn, hs.Sig, hs.Pipe)
	case CmdPayChanel:
		node.chargeServe(conn, hs.Sig, hs.Lic)
	}
}

func (node *PipeMiner) answerCheck(conn *JsonConn) {
	ack := YouPipeACK{Success: true}
	conn.WriteJsonMsg(ack)
	conn.Close()
}

func (node *PipeMiner) initCharger(conn *JsonConn, sig []byte, l *License) (*bandCharger, error) {

	if err := l.VerifySelf(sig); err != nil {
		logger.Warning("this license is not from real owner:%s", err)
		return nil, err
	}

	if err := l.VerifyData(); err != nil {
		logger.Warning("this license is not from our king:%s", err)
		return nil, err
	}

	if c := node.getCharger(l.UserAddr); c != nil {
		return nil, fmt.Errorf("duplicate payment channel")
	}

	charger := &bandCharger{
		JsonConn:   conn,
		receipt:    node.proofSaver.proofs,
		token:      BandWidthPerToPay,
		peerID:     account.ID(l.UserAddr),
		bill:       make(chan *PipeBill, MaxBandBill),
		done:       make(chan error),
		peerIPAddr: conn.RemoteAddr().String(),
	}

	acc := account.GetAccount()

	if err := acc.CreateAesKey(&charger.aesKey, l.UserAddr); err != nil {
		return nil, err
	}

	logger.Debug("create charger success:->", l.UserAddr)
	return charger, nil
}

func (node *PipeMiner) chargeServe(conn *JsonConn, sig []byte, l *License) {
	defer conn.Close()

	charger, err := node.initCharger(conn, sig, l)
	conn.writeAck(err)
	if err != nil {
		logger.Error(err)
		return
	}

	node.addCharger(charger)

	go charger.waitingReceipt()

	charger.charging()

	node.removeCharger(charger)

	logger.Info("charger exit:->", charger.peerID)

	return
}

func (node *PipeMiner) initPipe(sig []byte, req *PipeReqData) (net.Conn, *bandCharger, error) {

	if !req.Verify(sig) {
		return nil, nil, fmt.Errorf("signature failed")
	}

	conn, err := net.Dial("tcp", req.Target)
	if err != nil {
		return nil, nil, err
	}

	charger := node.getCharger(req.Addr)
	if charger == nil {
		return nil, nil, fmt.Errorf("no payment channel setup")
	}
	return conn, charger, nil
}

func (node *PipeMiner) pipeServe(conn *JsonConn, sig []byte, data *PipeReqData) {

	defer conn.Close()

	remoteConn, charger, err := node.initPipe(sig, data)
	conn.writeAck(err)
	if err != nil {
		logger.Warning(err)
		return
	}

	producerConn := NewProducerConn(conn.Conn, charger.aesKey)
	pipe := NewPipe(producerConn, remoteConn, charger, data.Target)
	logger.Infof("New pipe %s", pipe.String())

	go pipe.listenRequest()

	pipe.pushBackToClient()
}

func (node *PipeMiner) addCharger(c *bandCharger) {
	node.Lock()
	defer node.Unlock()
	node.chargers[c.peerID.ToString()] = c
	logger.Debugf("New Customer(%s)", c.peerID)
}

func (node *PipeMiner) getCharger(peerId string) *bandCharger {
	node.RLock()
	defer node.RUnlock()
	return node.chargers[peerId]
}

func (node *PipeMiner) removeCharger(c *bandCharger) {
	node.Lock()
	defer node.Unlock()
	delete(node.chargers, c.peerID.ToString())
	logger.Debugf("Remove Customer(%s)", c.peerID)
}

//TODO::
func (node *PipeMiner) PatrolLicense() {
}
