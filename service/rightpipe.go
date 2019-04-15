package service

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"io"
	"net"
	"sync"
	"time"
)

type PipeReqData struct {
	Addr   string
	Target string
}

type PipeRequest struct {
	Sig []byte
	*PipeReqData
}

func (a *PipeAdmin) addNewPipe(l net.Conn, target string) (*RightPipe, error) {
	r, err := net.Dial("tcp", target)
	if err != nil {
		err = fmt.Errorf("failed to connect Target %v", err)
		return nil, err
	}
	if err := r.(*net.TCPConn).SetKeepAlive(true); err != nil {
		return nil, err
	}

	l, err = Shadow(l, a.aesKey)
	if err != nil {
		err = fmt.Errorf("shadow the incoming conn err:%v", err)
		return nil, err
	}

	p := newPipe(l, r)
	a.Lock()
	defer a.Unlock()

	if _, ok := a.pipes[p.PipeID]; ok {
		err = fmt.Errorf("duplicate Target:%s", p.PipeID)
		return nil, err
	}

	a.pipes[p.PipeID] = p
	return p, nil
}

func (a *PipeAdmin) removePipe(pid string) {

	a.Lock()
	defer a.Unlock()

	p, ok := a.pipes[pid]
	if !ok {
		logger.Warning("no such pipe to remove:->", pid)
		return
	}
	p.close()
	delete(a.pipes, pid)
	logger.Debugf("remove pipe(%s)", p.PipeID)
}

type RightPipe struct {
	PipeID string
	err    error
	up     int64
	down   int64
	left   net.Conn
	right  net.Conn
}

func (p *RightPipe) pullFromServer() {
	n, err := io.Copy(p.right, p.left)
	p.up = n
	p.expire(err)
}

func (p *RightPipe) pushBackToClient() {
	n, err := io.Copy(p.left, p.right)
	p.down = n
	p.expire(err)
}

func (p *RightPipe) expire(err error) {
	p.right.SetDeadline(time.Now())
	p.left.SetDeadline(time.Now())
	if err == nil {
		return
	}

	if err, ok := err.(net.Error); !ok || !err.Timeout() {
		p.err = err
	}
}

func (p *RightPipe) close() {
	p.right.Close()
	logger.Debugf("pipe(%s) closing", p.PipeID)
}

func newPipe(l, r net.Conn) *RightPipe {
	pid := fmt.Sprintf("%s<->%s", l.RemoteAddr().String(), r.RemoteAddr().String())
	p := &RightPipe{
		PipeID: pid,
		left:   l,
		right:  r,
	}
	return p
}
