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
	_ PipeCmd = iota
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
	*PipeReqData
	*LicenseData
}

type PipeMiner struct {
	sync.RWMutex
	done       chan error
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

	node := &PipeMiner{
		done:       make(chan error),
		serverConn: s,
		chargers:   make(map[string]*bandCharger),
	}

	return node
}

func (node *PipeMiner) Mining() {
	defer node.serverConn.Close()
	for {
		conn, err := node.serverConn.Accept()
		if err != nil {
			panic(err)
		}

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
		node.pipeServe(conn, hs.Sig, hs.PipeReqData)
	case CmdPayChanel:
		node.chargeServe(conn, hs.Sig, hs.LicenseData)
	}
}

func (node *PipeMiner) answerCheck(conn *JsonConn) {
	ack := YouPipeACK{Success: true}
	conn.WriteJsonMsg(ack)
	conn.Close()
}

func (node *PipeMiner) initCharger(conn *JsonConn, sig []byte, data *LicenseData) (*bandCharger, error) {

	l := &License{
		Signature:   sig,
		LicenseData: data,
	}

	if err := l.Verify(); err != nil {
		return nil, err
	}

	if c := node.getCharger(l.UserAddr); c != nil {
		return nil, fmt.Errorf("duplicate payment channel")
	}

	charger := &bandCharger{
		JsonConn:   conn,
		token:      BandWidthPerToPay * 2,
		peerID:     account.ID(l.UserAddr),
		bill:       make(chan *PipeBill),
		receipt:    make(chan struct{}),
		peerIPAddr: conn.RemoteAddr().String(),
	}

	acc := account.GetAccount()

	if err := acc.CreateAesKey(&charger.aesKey, l.UserAddr); err != nil {
		return nil, err
	}

	return charger, nil
}

func (node *PipeMiner) chargeServe(conn *JsonConn, sig []byte, data *LicenseData) {
	defer conn.Close()

	charger, err := node.initCharger(conn, sig, data)
	conn.writeAck(err)
	if err != nil {
		logger.Error(err)
		return
	}

	node.addCharger(charger)

	node.done <- charger.charging()

	node.removeCharger(charger)

	return
}

func (node *PipeMiner) initPipe(sig []byte, data *PipeReqData) (net.Conn, *bandCharger, error) {

	req := &PipeRequest{
		Sig:         sig,
		PipeReqData: data,
	}

	if !req.Verify() {
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
		logger.Error(err)
		return
	}

	producerConn := NewProducerConn(conn, charger.aesKey)
	pipe := NewPipe(producerConn, remoteConn, charger)

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
