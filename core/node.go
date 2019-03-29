package core

import (
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/account"
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
	NodeId     string
	ServeNode  *service.SNode
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
		NodeId:     account.GetAccount().NodeId,
		GossipNode: gossip.GetGspNode(),
		ServeNode:  service.GetSNode(),
	}

	obj.GossipNode.NodeID = obj.NodeId
	obj.SetGspFilter()
	return obj
}

func (n *YouPipeNode) Start() {
	go n.GossipNode.JoinSwarm()
	go n.ServeNode.Mining()
}

func (n *YouPipeNode) Destroy() {
	n.GossipNode.Unsubscribe()
}
