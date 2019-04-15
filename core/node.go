package core

import (
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/gossip"
	"github.com/youpipe/go-youPipe/service"
	"github.com/youpipe/go-youPipe/utils"
	"sync"
)

var (
	instance  *YouPipeNode = nil
	once      sync.Once
	logger, _ = logging.GetLogger(utils.LMCore)
)

type YouPipeNode struct {
	ServeNode  *service.PipeMiner
	GossipNode *gossip.GNode
}

func GetNodeInst() *YouPipeNode {
	once.Do(func() {
		instance = newNode()
	})

	return instance
}

func newNode() *YouPipeNode {

	obj := &YouPipeNode{
		GossipNode: gossip.GetGspNode(),
		ServeNode:  service.GetMiner(),
	}

	obj.SetGspFilter()
	return obj
}

func (n *YouPipeNode) Start() {
	go n.GossipNode.JoinSwarm()
	go n.ServeNode.OpenPaymentChannel()
	go n.ServeNode.Mining()
}

func (n *YouPipeNode) Destroy() {
	n.GossipNode.Unsubscribe()
}
