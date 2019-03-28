package core

import (
	"fmt"
	"github.com/youpipe/go-youPipe/gossip"
	"github.com/youpipe/go-youPipe/pbs"
)

const (
	FindBootNode = pbs.YPMsgTyp_FindBootNode
	BootNodeAck  = pbs.YPMsgTyp_BootNodeAck
)

var (
	ETimeOut = fmt.Errorf("time out")
)

func (n *YouPipeNode) SetGspFilter() {
	_, err := n.GossipNode.Listen(&gossip.GspAddr{
		NodeId: gossip.BroadCastTarget,
		Type:   BootNodeAck,
	}, n.adminShowMe)

	if err != nil {
		logger.Warning("setup gossip connection err:->", err)
	}
}
