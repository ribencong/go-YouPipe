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
	serviceConn net.Listener
	users       map[string]*customer
}

func GetSNode() *SNode {
	once.Do(func() {
		instance = newServiceNode()
	})

	return instance
}

func newServiceNode() *SNode {
	id := account.GetAccount().Address

	addr := network.JoinHostPort(Config.ServiceIP, id.ToSocketPort())

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	logger.Infof("socks5 server listening TCP on %s", addr)

	node := &SNode{
		serviceConn: l,
		users:       make(map[string]*customer),
	}

	return node
}

func (node *SNode) Mining() {
	defer node.serviceConn.Close()

	for {
		conn, err := node.serviceConn.Accept()
		if err != nil {
			panic(err)
			return
		}

		node.newWaiter(conn).Start()
	}
}

func (node *SNode) getOrCreateCustomer(peerId string) *customer {
	node.Lock()
	defer node.Unlock()
	if u, ok := node.users[peerId]; ok {
		return u
	}

	user := &customer{
		address: peerId,
		pipes:   make(map[string]*Pipe),
	}

	if err := account.GetAccount().CreateAesKey(&user.aesKey, peerId); err != nil {
		return nil
	}

	node.users[peerId] = user
	return user
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
	logger.Debugf("remove user(%s)", peerId)

}
