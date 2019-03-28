package gossip

import (
	"fmt"
	"github.com/youpipe/go-node/utils"
	"net"
	"sync"
	"time"
)

type VNodeDirect int

const (
	OutDir VNodeDirect = 1
	InDir  VNodeDirect = 2
)

type viewNode struct {
	sync.RWMutex
	net.Conn
	dir         VNodeDirect
	peerID      string
	remoteIp    string
	probability float64
	HeartBeat   time.Time
}

func newInVNode(conn net.Conn, pid, rip string) *viewNode {

	node := &viewNode{
		Conn:      conn,
		dir:       InDir,
		peerID:    pid,
		remoteIp:  rip,
		HeartBeat: time.Now(),
	}

	return node
}

func newOutVNode(pid, rip string) (*viewNode, error) {

	addr := net.JoinHostPort(rip, ServicePort)

	conn, err := net.DialTimeout("tcp", addr, CommonTCPTimeOut)
	if err != nil {
		logger.Warning("create out view node err:->", err)
		return nil, err
	}
	conn.(*net.TCPConn).SetKeepAlive(true)

	node := &viewNode{
		Conn:      conn,
		dir:       OutDir,
		peerID:    pid,
		remoteIp:  rip,
		HeartBeat: time.Now(),
	}

	return node, nil
}

func (vn *viewNode) send(data []byte) error {
	vn.Lock()
	defer vn.Unlock()
	if _, err := vn.Write(data); err != nil {
		vn.Close()
		logger.Warningf("send to %s err:%s", vn.peerID, err)
		return err
	}

	vn.HeartBeat = time.Now()
	return nil
}

func (vn *viewNode) Update() {
	vn.Lock()
	defer vn.Unlock()
	vn.HeartBeat = time.Now()
}
func (vn *viewNode) Destroy() {
	vn.Lock()
	defer vn.Unlock()
	vn.Close()
}

func (vn *viewNode) updateWeight(wei float64) {
	vn.Lock()
	defer vn.Unlock()
	vn.probability = wei
}

func (vn *viewNode) String() string {
	vn.RLock()
	defer vn.RUnlock()

	str := fmt.Sprintf("\n||||||||||||||||||||||||||||||||||||||||||||||||\n"+
		"|%s|\n"+
		"|%-15s:%30s|\n"+
		"|%-15s:%30.2f|\n"+
		"|%-15s:%30s|\n"+
		"|%-15s:%30d|\n"+
		"||||||||||||||||||||||||||||||||||||||||||||||||",
		vn.peerID, "remoteIP", vn.remoteIp,
		"probability", vn.probability,
		"HeartBeat", vn.HeartBeat.Format(utils.SysTimeFormat),
		"direction", vn.dir)

	return str
}
