package service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/youpipe/go-youPipe/pbs"
	"github.com/youpipe/go-youPipe/thread"
	"io"
	"net"
	"time"
)

type connWaiter struct {
	*customer
	net.Conn
	*SNode
}

func (node *SNode) newWaiter(conn net.Conn) *thread.Thread {
	rAddr := conn.RemoteAddr().String()

	w := &connWaiter{
		Conn:  conn,
		SNode: node,
		customer: &customer{
			remoteAddr: rAddr,
		},
	}

	t := thread.NewThreadWithName(w, rAddr)
	logger.Debugf("new customer come in :->", t.Name)
	return t
}

func (cw *connWaiter) CloseCallBack(t *thread.Thread) {
}

func (cw *connWaiter) DebugInfo() string {
	return fmt.Sprintf("\n||||||||||||||||||||||||||||||||||||||||||||||||\n"+
		"|%s|\n"+
		"|%-15s:%30s|\n"+
		"||||||||||||||||||||||||||||||||||||||||||||||||",
		cw.accountId,
		"remoteIP", cw.remoteAddr)
}

func (cw *connWaiter) Run(ctx context.Context) {

	defer cw.Close()
	cw.Conn.(*net.TCPConn).SetKeepAlive(true)

	addr, err := handShake(cw)
	if err != nil {
		logger.Warningf("failed to read address:->%v", err)
		return
	}

	remoteConn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.Warningf("failed to connect target:->%v", err)
		return
	}
	defer remoteConn.Close()

	remoteConn.(*net.TCPConn).SetKeepAlive(true)
	logger.Infof("proxy %s <-> %s", cw.remoteAddr, addr)

	_, _, err = relay(cw, remoteConn)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return // ignore i/o timeout
		}
		logger.Warningf("relay error: %v", err)
	}
}

func handShake(conn net.Conn) (string, error) {
	buffer := make([]byte, MaxAddrLen)
	n, err := conn.Read(buffer)
	if err != nil {
		logger.Warningf("failed to read address:->%v", err)
		return "", err
	}

	addr := &pbs.Sock5Addr{}
	if err := proto.Unmarshal(buffer[:n], addr); err != nil {
		logger.Warningf("unmarshal address:->%v", err)
		return "", err
	}

	res, _ := proto.Marshal(&pbs.CommAck{
		ErrNo:  pbs.ErrorNo_Success,
		ErrMsg: "success",
	})

	if _, err := conn.Write(res); err != nil {
		logger.Warningf("write response err :->%v", err)
		return "", err
	}

	return net.JoinHostPort(addr.Host, addr.Port), nil
}

func relay(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		n, err := io.Copy(right, left)
		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
		ch <- res{n, err}
	}()

	n, err := io.Copy(left, right)
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}
	return n, rs.N, err
}
