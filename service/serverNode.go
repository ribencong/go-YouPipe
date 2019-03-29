package service

import (
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/utils"
	"net"
	"sync"
)

var (
	instance  *SNode = nil
	once      sync.Once
	logger, _ = logging.GetLogger(utils.LMService)
)

const MaxAddrLen = 1 + 1 + 255 + 2

type SNode struct {
	sync.RWMutex
	serviceConn net.Listener
}

func GetSNode() *SNode {
	once.Do(func() {
		instance = newServiceNode()
	})

	return instance
}

func newServiceNode() *SNode {

	addr := Config.ServicePoint
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	logger.Infof("socks5 server listening TCP on %s", addr)

	node := &SNode{
		serviceConn: l,
	}

	return node
}

func (node *SNode) Mining() {
	defer node.serviceConn.Close()

	for {
		conn, err := node.serviceConn.Accept()
		if err != nil {
			logger.Warningf("failed to accept :->%v", err)
			continue
		}

		node.newWaiter(conn).Start()
	}
}
