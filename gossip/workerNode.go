package gossip

import (
	"context"
	"fmt"
	"github.com/ribencong/go-youPipe/pbs"
	"github.com/ribencong/go-youPipe/thread"
	"math/rand"
	"net"
)

type workNode struct {
	*viewNode
	*GNode
}

func newWorkerNode(n *GNode, vn *viewNode) *thread.Thread {
	s := &workNode{
		GNode:    n,
		viewNode: vn,
	}

	t := thread.NewThread(s)
	t.Name = fmt.Sprintf("%d-%s", vn.dir, s.peerID)
	logger.Debugf("worker node(%s) with ip(%s) created thread(%d) ", s.peerID, s.remoteIp, t.ID)
	return t
}

func (wn *workNode) CloseCallBack(t *thread.Thread) {
	if wn.dir == OutDir {
		wn.outPut.Remove(wn.peerID)
	} else {
		wn.income.Remove(wn.peerID)
	}
}

func (wn *workNode) DebugInfo() string {
	return wn.viewNode.String()
}

func (wn *workNode) Run(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			logger.Warning("thread closed by other")
			return
		default:
			typ, msg, err := pullMsg(wn)
			if err != nil {
				logger.Warning("worker node exit for error msg:->", err, wn.peerID)
				return
			}
			if err := wn.process(typ, msg); err != nil {
				logger.Warning("parse control message err:->", err)
				continue
			}
		}
	}
}

func (wn *workNode) process(t pbs.MsgType, msg *pbs.Gossip) error {

	pAddr := wn.RemoteAddr().String()
	logger.Infof("from %s typ(%s) message:%s ", pAddr, t, msg)

	switch int(t) {
	case HeartBeat:
		return wn.gotHB(msg.ID.NodeId)

	case UpdateWeight:
		w := msg.UpdateWeight
		dir := VNodeDirect(w.Direct)

		logger.Debugf("update weight(%.2f) of node(%s)"+
			" in my (%d) view", w.Weight, w.NodeId, dir)

		if dir == InDir {
			return wn.outPut.updateNodeWeight(w.NodeId, w.Weight)
		} else {
			return wn.income.updateNodeWeight(w.NodeId, w.Weight)
		}

	case ReSubscribe:
		nodeId := msg.ID.NodeId
		ip, _, _ := net.SplitHostPort(pAddr)
		ttl := wn.outPut.Size() * 2
		return wn.voteOrAccept(nodeId, ip, ttl)

	case VoteContact:
		vote := msg.Vote
		return wn.voteOrAccept(vote.NodeId, vote.IP, int(vote.TTL))

	case Forward:
		return wn.goOnOrWelcome(msg.Forward)

	case ReplaceView:
		return wn.replaceUnSubNode(msg.RplView)

	case AppPayload:
		return wn.appMsgGot(msg.AppMsg)
	default:
		return EInvalidMsg
	}
}

func (n *GNode) replaceUnSubNode(rp *pbs.Replace) error {

	if n.NodeID == rp.AlterId {
		return ESelfReq
	}

	if n.outPut.Has(rp.AlterId) {
		return EDuplicateConn
	}

	logger.Debug("replace by a new one :->", rp.AlterId, rp.IP)
	node, err := newOutVNode(rp.AlterId, rp.IP)
	if err != nil {
		logger.Warning("output node  create err:->", err)
		return err
	}
	data, _ := pack(NewForReplace, n.NodeID)
	if err := node.send(data); err != nil {
		logger.Warning("notify peer's peer for replacement err:->", err)
		return err
	}
	n.outPut.Add(node, n.NodeID)
	newWorkerNode(n, node).Start()
	return nil
}

func (n *GNode) gotHB(nodeId string) error {

	node, ok := n.income.Get(nodeId)
	if !ok {
		return ENotFound
	}

	node.Update()

	return nil
}

func (n *GNode) goOnOrWelcome(msg *pbs.ForwardMsg) error {
	nodeId := msg.NodeId

	n.counter.Inc(msg.MsgId)
	if n.counter.IsOverFlood(msg.MsgId) {
		logger.Info(EOverForward, msg.MsgId)
		return EOverForward
	}

	prob := float64(1) / float64(1+n.outPut.Size())
	randProb := rand.Float64()

	if t := n.outPut.Has(nodeId); t || randProb > prob || nodeId == n.NodeID {
		logger.Debugf("Can't keep(%s,%s) because "+
			"prb=%.2f rand=%.2f duplicate=%t isSelf=%t",
			msg.IP, nodeId,
			prob, randProb, t, nodeId == n.NodeID)
		data, _ := pack(Forward, msg.NodeId, msg.IP, msg.MsgId)
		return n.outPut.RandomSend(data)
	}

	return n.acceptForwardSub(msg)
}
