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
	sync.RWMutex
	serviceConn  net.Listener
	microPayConn net.Listener
	users        map[string]*service
}

func GetMiner() *PipeMiner {
	once.Do(func() {
		instance = newMiner()
	})

	return instance
}

func newMiner() *PipeMiner {
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

	node := &PipeMiner{
		serviceConn:  l,
		microPayConn: p,
		users:        make(map[string]*service),
	}

	return node
}

func (node *PipeMiner) OpenPaymentChannel() {
	defer node.microPayConn.Close()
	for {
		conn, err := node.microPayConn.Accept()
		if err != nil {
			panic(err)
		}
		c := &JsonConn{conn}
		go node.newCustomer(c)
	}
}

func (node *PipeMiner) Mining() {
	defer node.serviceConn.Close()

	for {
		conn, err := node.serviceConn.Accept()
		if err != nil {
			panic(err)
		}
		c := &JsonConn{conn}
		go node.handShake(c)
	}
}

func (node *PipeMiner) addCustomer(peerID string, user *service) {
	node.Lock()
	defer node.Unlock()
	node.users[peerID] = user
	logger.Debugf("New Customer(%s)", peerID)
}

func (node *PipeMiner) getCustomer(peerId string) *service {
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
