package gossip

import (
	"github.com/youpipe/go-youPipe/pbs"
	"github.com/youpipe/go-youPipe/utils"
	"net"
	"time"
)

func (n *GNode) processNewConn(conn net.Conn) {
	t, msg, err := pullMsg(conn)
	if err != nil {
		logger.Warning("pull msg err:->", err)
		return
	}

	rAddr := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(rAddr)

	logger.Infof("new msg(%s) from (%s) msg(%s)", t, rAddr, msg)

	switch int(t) {

	case SubInit:
		ttl := n.outPut.Size() * 2
		err = n.voteOrAccept(msg.ID.NodeId, ip, ttl)
		conn.Close()

	case WelCome, SubSuccess, NewForReplace:
		nodeId := msg.ID.NodeId
		n.beWelcomed(conn, nodeId, ip)

	case GotContact:
		nodeId := msg.IDWithIP.NodeId
		n.beWelcomed(conn, nodeId, ip)
		err = n.foundContact(ip, nodeId, msg.IDWithIP.IP)

	default:
		logger.Warning(EInvalidMsg, msg)
	}

	if err != nil {
		logger.Warning("process new conn err:->", err)
	}
}

func (n *GNode) beWelcomed(conn net.Conn, nodeId, rIP string) {

	if _, ok := n.income.Get(nodeId); ok {
		logger.Debugf("I'm welcomed by %s repeatedly", nodeId)
		conn.Close()
		return
	}

	if n.expired.IsZero() {
		n.expired = time.Now().Add(Config.GspSubDuration)
		logger.Info("expire clock started", n.expired.Format(utils.SysTimeFormat))
	}

	nodeIn := newInVNode(conn, nodeId, rIP)
	n.income.Add(nodeIn, n.NodeID)
	newWorkerNode(n, nodeIn).Start()
}

func (n *GNode) foundContact(rIP, peerID, myIp string) error {

	if _, ok := n.outPut.Get(peerID); ok {
		logger.Debugf("I'm noticed by %s repeatedly", peerID)
		return nil
	}

	nodeOut, err := newOutVNode(peerID, rIP)
	if err != nil {
		return err
	}

	n.VisibleIp = myIp

	n.outPut.Add(nodeOut, n.NodeID)
	newWorkerNode(n, nodeOut).Start()

	data, _ := pack(SubSuccess, n.NodeID)
	return nodeOut.send(data)
}

func (n *GNode) voteOrAccept(nid, ip string, ttl int) (err error) {

	if n.NodeID == nid {
		n.VisibleIp = ip
		return ESelfReq
	}

	if ttl > 0 && n.outPut.Size() > 0 {

		data, _ := pack(VoteContact, nid, ip, int32(ttl-1))
		logger.Debug("Vote a real contact for you :->", nid, ip)
		if err = n.outPut.ChoseByProb(data); err == nil {
			return nil
		}
		logger.Warning("contact you because Vote err:->", nid, ip, err)
	}

	if node, ok := n.outPut.Get(nid); ok {
		return n.asContacts(nid, ip, node)
	}

	nodeOut, err := newOutVNode(nid, ip)
	if err != nil {
		return err
	}

	if err := n.asContacts(nid, ip, nodeOut); err != nil {
		return err
	}

	n.outPut.Add(nodeOut, n.NodeID)
	newWorkerNode(n, nodeOut).Start()
	return nil
}

func (n *GNode) acceptForwardSub(msg *pbs.ForwardMsg) error {

	logger.Debug("accept forwarded :->", msg.NodeId, msg.IP)
	node, err := newOutVNode(msg.NodeId, msg.IP)
	if err != nil {
		return err
	}

	n.outPut.Add(node, n.NodeID)
	newWorkerNode(n, node).Start()

	data, _ := pack(WelCome, n.NodeID)
	return node.send(data)
}

func (n *GNode) asContacts(nid, ip string, node *viewNode) error {

	logger.Debugf("notify new node(%s, %s) to online", nid, ip)
	data, _ := pack(GotContact, n.NodeID, node.remoteIp)
	if err := node.send(data); err != nil {
		logger.Warning("notify init sub err:->", err, nid, ip)
		return err
	}

	msgId := GenMsgID(n.NodeID)
	data, _ = pack(Forward, nid, ip, msgId)
	n.outPut.Broadcast(data)

	for i := 0; i < ConditionalForward; i++ {
		if err := n.outPut.RandomSend(data); err != nil {
			logger.Warning("conditional send err:->", err)
		}
	}

	return nil
}
