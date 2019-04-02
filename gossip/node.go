package gossip

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/utils"
	"net"
	"sync"
	"time"
)

var (
	instance  *GNode = nil
	once      sync.Once
	logger, _ = logging.GetLogger(utils.LMGossip)
)

func GetGspNode() *GNode {
	once.Do(func() {
		instance = newGossipNode()
	})

	return instance
}

type GNode struct {
	NodeID    string
	VisibleIp string
	expired   time.Time

	server net.Listener
	timers TimerTask

	counter *counter
	income  *ViewCache
	outPut  *ViewCache

	gcm *gspConnMgn
}

func newGossipNode() *GNode {
	l, err := net.Listen("tcp", ":"+ServicePort)
	if err != nil {
		panic(err)
	}
	logger.Infof("gossip server listening(%s)", l.Addr().String())
	obj := &GNode{
		NodeID:  account.GetAccount().Address,
		server:  l,
		timers:  make(TimerTask),
		income:  newCache("IN"),
		outPut:  newCache("OUT"),
		counter: newCounter(),
		gcm:     newConnMgn(),
	}

	obj.timers.Add(HeartBeatTime, obj.heartBeatTimer, "gossipHeartBeat")
	obj.timers.Add(RetrySubInterval, obj.isolateCheck, "gossipIsolateCheck")

	return obj
}

func (n *GNode) subscribe() {

	bootAddr := net.JoinHostPort(network.Config.BootStrapServer, ServicePort)
	data, _ := pack(SubInit, n.NodeID)

	logger.Debugf("<------Booting %s------>", bootAddr)
	conn, err := net.DialTimeout("tcp", bootAddr, CommonTCPTimeOut)
	if err != nil {
		logger.Warning("boot failed:->", err)
		return
	}
	defer conn.Close()

	if _, err := conn.Write(data); err != nil {
		logger.Warning("initial subscribe err:->", err)
	}
}

func (n *GNode) JoinSwarm() {

	n.timers.StartAll()

	n.subscribe()

	for {
		conn, err := n.server.Accept()
		if err != nil {
			logger.Warning("warning: gossip error accept:->", err)
			panic(err)
		}

		go n.processNewConn(conn)
	}
}

func (n *GNode) String() string {
	str := fmt.Sprintf("\n\n------%s------"+
		n.outPut.String("OUT")+
		n.income.String("IN ")+
		"\n---expire:%s-----%15s---------\n",
		n.NodeID,
		n.expired.Format(utils.SysTimeFormat),
		n.VisibleIp)
	return str
}

func (n *GNode) Unsubscribe() {

	outPutIds := n.outPut.AllKeys()
	inPutIds := n.income.AllKeys()
	lenIn := len(inPutIds)
	lenOut := len(outPutIds)

	logger.Debugf("system exit and unsubscribe to other nodes lenIn=%d, lenOut=%d", lenIn, lenOut)

	for idx := 0; idx < lenIn-ConditionalForward-1; idx++ {

		jIdx := idx % lenOut
		alterId := outPutIds[jIdx]
		inId := inPutIds[idx]

		outNode, ok := n.outPut.Get(alterId)
		if !ok {
			logger.Warningf("the altered node(%s) is not available any more", alterId)
			continue
		}

		logger.Debugf("tell %s to replace me with %s", inId, alterId)
		data, _ := pack(ReplaceView, n.NodeID, outNode.peerID, outNode.remoteIp)
		inNode, ok := n.income.Get(inId)

		if !ok {
			logger.Warningf("the in node(%s) doesn't exist", inId)
			continue
		}

		if err := inNode.send(data); err != nil {
			logger.Warning("failed to notify in node about my unsubscription:->", err)
		}
	}
}
