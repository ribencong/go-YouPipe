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
	instance  *SNode = nil
	once      sync.Once
	logger, _ = logging.GetLogger(utils.LMService)
)

type SNode struct {
	sync.RWMutex
	serviceConn  net.Listener
	microPayConn net.Listener
	users        map[string]*customer
}

func GetSNode() *SNode {
	once.Do(func() {
		instance = newServiceNode()
	})

	return instance
}

func newServiceNode() *SNode {
	id := account.GetAccount().Address
	servicePort := id.ToSocketPort()
	addr := network.JoinHostPort(Config.ServiceIP, servicePort)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	logger.Infof("Socks5 server listening TCP on %s", addr)

	payPort := servicePort + 1
	addr = network.JoinHostPort(Config.ServiceIP, payPort)
	p, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	logger.Infof("MicroPayment chanel starting at: %s", addr)

	node := &SNode{
		serviceConn:  l,
		microPayConn: p,
		users:        make(map[string]*customer),
	}

	return node
}

func (node *SNode) OpenPaymentChannel() {
	defer node.microPayConn.Close()
	for {
		conn, err := node.microPayConn.Accept()
		if err != nil {
			panic(err)
		}
		c := &CtrlConn{conn}
		go node.newCustomer(c)
	}
}

func (node *SNode) Mining() {
	defer node.serviceConn.Close()

	for {
		conn, err := node.serviceConn.Accept()
		if err != nil {
			panic(err)
		}
		c := &CtrlConn{conn}
		go node.handShake(c)
	}
}

func (node *SNode) addCustomer(peerID string, user *customer) {
	node.Lock()
	defer node.Unlock()
	node.users[peerID] = user
	logger.Debugf("New Customer(%s)", peerID)
}

func (node *SNode) getCustomer(peerId string) *customer {
	node.RLock()
	defer node.RUnlock()
	return node.users[peerId]
}

func (node *SNode) removeUser(peerId string) {
	node.Lock()
	defer node.Unlock()
	delete(node.users, peerId)
	logger.Debugf("Remove Customer(%s)", peerId)
}
