package service

import (
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

type PipeMiner struct {
	*account.Account
	serverConn net.Listener
	sync.RWMutex
	users map[string]*customer
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
		Account:    acc,
		serverConn: s,
		users:      make(map[string]*customer),
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
	case CmdPayChanel:
		node.createCharger(conn, hs.Sig, hs.LicenseData)
	}
}

func (node *PipeMiner) answerCheck(conn *JsonConn) {
	ack := YouPipeACK{Success: true}
	conn.WriteJsonMsg(ack)
	conn.Close()
}
func (node *PipeMiner) createCharger(conn *JsonConn, sig []byte, data *LicenseData) error {

	l := &License{
		Signature:   sig,
		LicenseData: data,
	}
	if err := l.Verify(); err != nil {
		conn.writeAck(err)
	}

	charger := &BWCharger{
		JsonConn: conn,
		priKey:   node.Account.Key.PriKey,
	}

	charger.monitoPipe()

	return nil
}

func (node *PipeMiner) addCustomer(peerID string, user *customer) {
	node.Lock()
	defer node.Unlock()
	node.users[peerID] = user
	logger.Debugf("New Customer(%s)", peerID)
}

func (node *PipeMiner) getCustomer(peerId string) *customer {
	node.RLock()
	defer node.RUnlock()
	return node.users[peerId]
}

func (node *PipeMiner) removeUser(peerId string) {
	node.Lock()
	defer node.Unlock()
	delete(node.users, peerId)
	logger.Debugf("Remove Customer(%s)", peerId)
}
