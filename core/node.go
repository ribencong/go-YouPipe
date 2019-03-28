package core

import (
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/gossip"
	"github.com/youpipe/go-youPipe/service"
	"github.com/youpipe/go-youPipe/utils"
	"sync"
)

var (
	instance *YouPipeNode = nil
	once     sync.Once
	logger   = utils.NewLog(utils.LMCore)
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

	logger.Info("<---Create YouPipe Node Success--->")
	return obj
}

func (n *YouPipeNode) Run() {
	go n.GossipNode.JoinSwarm()
	go n.ServeNode.Mining()
	logger.Info("<---YouPipe node start working--->")
}

func (n *YouPipeNode) Destroy() {
	n.GossipNode.Unsubscribe()
}
