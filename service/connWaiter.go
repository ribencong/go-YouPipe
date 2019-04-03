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
	*pbs.Sock5Req
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
	logger.Debugf("new customer come in :->", t.Name)
	return t
}

func (cw *connWaiter) CloseCallBack(t *thread.Thread) {
	cw.Close()

	if u := cw.getCustomer(cw.Address); u != nil {
		u.removePipe(cw.Target)

		if u.isPipeEmpty() {
			cw.removeUser(cw.Address)
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
	if err := cw.handShake(); err != nil {
		logger.Warningf("failed to parse socks5 request:->%v", err)
		return
	}

	user := cw.getOrCreateCustomer(cw.Address)
	if nil == user {
		logger.Warning("get customer info err:->", cw.Target, cw.Address)
		return
	}

	pipe := user.addNewPipe(cw, cw.Target)
	if pipe == nil {
		logger.Warning("create new pipe failed:->")
		return
	}
	go pipe.pull()

	pipe.push()

	logger.Warning("pipe(up=%d, down=%d) is broken err=%v:->", pipe.up, pipe.down, pipe.err)
}

func (cw *connWaiter) handShake() error {
	buffer := make([]byte, buffSize)
	n, err := cw.Read(buffer)
	if err != nil {
		logger.Warningf("failed to read address:->%v", err)
		return err
	}

	sockReq := &pbs.Sock5Req{}
	if err := proto.Unmarshal(buffer[:n], sockReq); err != nil {
		logger.Warningf("unmarshal address:->%v", err)
		return err
	}
	myId := account.GetAccount().Address
	res, _ := proto.Marshal(&pbs.Sock5Res{
		Address: string(myId),
	})

	if _, err := cw.Write(res); err != nil {
		logger.Warningf("write response err :->%v", err)
		return err
	}
	cw.Sock5Req = sockReq

	return nil
}

//func relay(left, right net.Conn) (int64, int64, error) {
//	type res struct {
//		N   int64
//		Err error
//	}
//	ch := make(chan res)
//
//	go func() {
//		n, err := io.Copy(right, left)
//		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
//		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
//		ch <- res{n, err}
//	}()
//
//	n, err := io.Copy(left, right)
//	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
//	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
//	rs := <-ch
//
//	if err == nil {
//		err = rs.Err
//	}
//	return n, rs.N, err
//}
