package gossip

import (
	"fmt"
	"github.com/ribencong/go-youPipe/pbs"
	"sync"
)

/************************************************************************
*			gossip address used by gossip connection
************************************************************************/
type GspAddr struct {
	NodeId string
	Type   pbs.YPMsgTyp
}

//func SplitGspAddr(addr string) (*GspAddr, error) {
//
//	str := strings.Split(addr, ":")
//
//	if len(str) != 2 {
//		return nil, EAppAddr
//	}
//
//	t, err := strconv.Atoi(str[1])
//	if err != nil {
//		return nil, err
//	}
//
//	return &GspAddr{
//		Address: str[0],
//		Type:   int32(t),
//	}, nil
//}

func (ga *GspAddr) Join() string {
	return fmt.Sprintf("%s:%d", ga.NodeId, ga.Type)
}

func (ga *GspAddr) BCJoin() string {
	return fmt.Sprintf("%s:%d", BroadCastTarget, ga.Type)
}

func (ga *GspAddr) UCJoin(id string) string {
	return fmt.Sprintf("%s:%d", id, ga.Type)
}

func (ga *GspAddr) IsBroadCast(id string) bool {
	return ga.NodeId == BroadCastTarget
}

/************************************************************************
*			gossip connection used by high level
************************************************************************/
type ConnHandler func(*GspConn, *pbs.AppMsg) error
type GspConn struct {
	*GNode
	connected bool
	LAddr     *GspAddr
	RAddr     *GspAddr
	watch     ConnHandler
}

func (conn *GspConn) SendBySize(data []byte) error {
	ttl := conn.outPut.Size() * 2
	return conn.SendByTTL(data, int32(ttl))
}
func (conn *GspConn) SendToBySize(data []byte, dst string) error {
	ttl := conn.outPut.Size() * 2
	return conn.SendToByTTL(data, int32(ttl), dst)
}

func (conn *GspConn) Send(data []byte) error {
	return conn.SendByTTL(data, AppMsgBroadCast)
}
func (conn *GspConn) SendTo(data []byte, dst string) error {
	return conn.SendToByTTL(data, AppMsgBroadCast, dst)
}

func (conn *GspConn) SendByTTL(data []byte, ttl int32) error {
	if conn.RAddr == nil {
		return ENotConnected
	}
	return conn.SendToByTTL(data, ttl, conn.RAddr.Join())
}
func (conn *GspConn) SendToByTTL(data []byte, ttl int32, dst string) error {

	msgId := GenMsgID(conn.NodeID)
	msgData, _ := pack(AppPayload, msgId, conn.LAddr.Join(),
		dst, ttl, data)

	logger.Debug("send app msg by ID:->", msgId)
	conn.counter.Inc(msgId)
	conn.outPut.Broadcast(msgData)
	return nil
}

/************************************************************************
*			gossip connection manager
************************************************************************/
type gspConnMgn struct {
	sync.RWMutex
	watching map[string]*GspConn
}

func newConnMgn() *gspConnMgn {
	return &gspConnMgn{
		watching: make(map[string]*GspConn),
	}
}

func (conn *GspConn) Finish() {
	conn.gcm.Lock()
	defer conn.gcm.Unlock()
	delete(conn.gcm.watching, conn.LAddr.Join())
}

func (n *GNode) addFilter(la, ra *GspAddr, w ConnHandler) (*GspConn, error) {
	id := la.Join()

	n.gcm.Lock()
	defer n.gcm.Unlock()

	if c, ok := n.gcm.watching[id]; ok {
		return c, EInUsed
	}

	c := &GspConn{
		GNode:     n,
		connected: ra != nil,
		LAddr:     la,
		RAddr:     ra,
		watch:     w,
	}
	n.gcm.watching[id] = c
	return c, nil
}

func (gcm *gspConnMgn) filter(ra string) *GspConn {

	gcm.RLock()
	defer gcm.RUnlock()
	c, ok := gcm.watching[ra]

	if !ok {
		return nil
	}
	return c
}

/************************************************************************
*		gossip node process application message internal
************************************************************************/

func (n *GNode) relay(appMsg *pbs.AppMsg) {

	if appMsg.TTL--; appMsg.TTL <= 0 {
		logger.Debug("msg is out of live:->", appMsg)
		return
	}

	logger.Debug("relay:->", appMsg)
	data, _ := pack(AppPayload, appMsg.MsgId, appMsg.LAddr,
		appMsg.RAddr, appMsg.TTL, appMsg.PayLoad)
	n.outPut.Broadcast(data)
}

func (n *GNode) appMsgGot(appMsg *pbs.AppMsg) error {

	msgID := appMsg.MsgId
	if n.counter.Has(msgID) {
		logger.Debug("duplicated :", msgID)
		return nil
	}
	n.counter.Inc(msgID)

	conn := n.gcm.filter(appMsg.RAddr)
	if conn == nil {
		n.relay(appMsg)
		return nil
	}

	logger.Debugf("night watcher(isCon=%t, w=%s) for %s",
		conn.connected, conn.LAddr.Join(), msgID)
	if err := conn.watch(conn, appMsg); err != nil {
		return err
	}

	if conn.connected {
		logger.Debugf("it's only for me:%s", msgID)
		return nil
	}

	n.relay(appMsg)
	return nil
}

/************************************************************************
*		gossip node external API
************************************************************************/
func (n *GNode) Dial(la, ra *GspAddr, w ConnHandler) (*GspConn, error) {
	return n.addFilter(la, ra, w)
}

func (n *GNode) Listen(addr *GspAddr, w ConnHandler) (*GspConn, error) {
	return n.addFilter(addr, nil, w)
}
