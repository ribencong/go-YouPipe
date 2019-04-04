package service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/pbs"
	"github.com/youpipe/go-youPipe/thread"
	"net"
)

type connWaiter struct {
	customerId string
	pipeId     string
	net.Conn
	*SNode
}

func (node *SNode) newWaiter(conn net.Conn) *thread.Thread {
	conn.(*net.TCPConn).SetKeepAlive(true)

	rAddr := conn.RemoteAddr().String()
	w := &connWaiter{
		Conn:  conn,
		SNode: node,
	}

	t := thread.NewThreadWithName(w, rAddr)
	logger.Debugf("new customer(%s) coming.", t.Name)
	return t
}

func (cw *connWaiter) CloseCallBack(t *thread.Thread) {
	cw.Close()
	if len(cw.customerId) == 0 {
		return
	}

	if u := cw.getCustomer(cw.customerId); u != nil {

		u.removePipe(cw.pipeId)

		if u.isPipeEmpty() {
			cw.removeUser(cw.customerId)
		}
	}
}

func (cw *connWaiter) DebugInfo() string {
	return fmt.Sprintf("\n||||||||||||||||||||||||||||||||||||||||||||||||\n"+
		//"|%s|\n"+
		"|%-15s:%30s|\n"+
		"||||||||||||||||||||||||||||||||||||||||||||||||",
		//cw.RemoteAddr().String(),
		"remoteIP", cw.RemoteAddr().String())
}

func (cw *connWaiter) Run(ctx context.Context) {

	req, err := cw.handShake()
	if err != nil {
		logger.Warningf("failed to parse socks5 request:->%v", err)
		return
	}
	logger.Debug("get request:", req)
	user := cw.getOrCreateCustomer(req.Address)
	if nil == user {
		logger.Warning("get customer info err:->", req)
		return
	}

	pipe := user.addNewPipe(cw.Conn, req.Target, req.IsRaw)
	if pipe == nil {
		logger.Warning("create new pipe failed:->")
		return
	}
	cw.pipeId = pipe.PipeID

	logger.Debugf("proxy %s <-> %s", cw.RemoteAddr().String(), req.Target)

	go pipe.pull()

	pipe.push()

	logger.Warningf("pipe(%s) is broken(up=%d, down=%d) err=%v:", pipe.PipeID, pipe.up, pipe.down, pipe.err)
}

func (cw *connWaiter) handShake() (*pbs.Sock5Req, error) {
	buffer := make([]byte, buffSize)
	n, err := cw.Read(buffer)
	if err != nil {
		logger.Warningf("failed to read address:->%v", err)
		return nil, err
	}

	sockReq := &pbs.Sock5Req{}
	if err := proto.Unmarshal(buffer[:n], sockReq); err != nil {
		logger.Warningf("unmarshal address:->%v", err)
		return nil, err
	}
	myId := account.GetAccount().Address
	res, _ := proto.Marshal(&pbs.Sock5Res{
		Address: string(myId),
	})

	if _, err := cw.Write(res); err != nil {
		logger.Warningf("write response err :->%v", err)
		return nil, err
	}
	cw.customerId = sockReq.Address
	return sockReq, nil
}
