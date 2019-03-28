package core

import (
	"github.com/golang/protobuf/proto"
	"github.com/youpipe/go-node/gossip"
	"github.com/youpipe/go-node/pbs"
	"github.com/youpipe/go-node/thread"
	"time"
)

func (n *YouPipeNode) AdminFindBootStrap(maxSize int32) ([]*pbs.BootNodes, error) {

	res := make(chan *pbs.YouPipeMsg, maxSize)

	callBack := func(conn *gossip.GspConn, appMsg *pbs.AppMsg) error {
		msg := &pbs.YouPipeMsg{}
		if err := proto.Unmarshal(appMsg.PayLoad, msg); err != nil {
			return err
		}
		res <- msg
		return nil
	}

	conn, err := n.GossipNode.Dial(&gossip.GspAddr{
		NodeId: n.NodeId,
		Type:   FindBootNode,
	}, &gossip.GspAddr{
		NodeId: gossip.BroadCastTarget,
		Type:   BootNodeAck,
	}, callBack)

	if err != nil {
		return nil, err
	}

	defer conn.Finish()

	logger.Debugf("send gossip conn msg %s<-> %s", conn.LAddr.Join(), conn.RAddr.Join())
	data, _ := proto.Marshal(&pbs.YouPipeMsg{
		Typ: FindBootNode,
	})

	if err := conn.SendBySize(data); err != nil {
		return nil, err
	}
	nodes := make([]*pbs.BootNodes, 0)
	for {

		select {
		case <-time.After(time.Second * 4):
			return nodes, ETimeOut

		case msg := <-res:
			maxSize--
			nodes = append(nodes, msg.Nodes)
			logger.Debug("got one node:->", maxSize, nodes)
			if maxSize <= 0 {
				return nodes, nil
			}
		}
	}
}

func (n *YouPipeNode) adminShowMe(conn *gossip.GspConn, appMsg *pbs.AppMsg) error {

	msg := &pbs.YouPipeMsg{}
	if err := proto.Unmarshal(appMsg.PayLoad, msg); err != nil {
		return err
	}

	logger.Debug("got require :->", msg, appMsg.LAddr, conn.LAddr.Join())

	payLoad := thread.ThreadNO()
	res := pbs.YouPipeMsg{
		Typ: BootNodeAck,
		Nodes: &pbs.BootNodes{
			NodeId:  n.NodeId,
			PeerIP:  n.GossipNode.VisibleIp,
			PayLoad: int32(payLoad),
		},
	}

	logger.Debug("and answer with:->", res)

	data, _ := proto.Marshal(&res)
	return conn.SendTo(data, appMsg.LAddr)
}
