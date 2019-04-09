package service

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Pipe struct {
	PipeID string
	err    error
	up     int64
	down   int64
	left   net.Conn
	right  net.Conn
}

func (p *Pipe) pull() {
	n, err := io.Copy(p.right, p.left)
	p.up = n
	p.expire(err)
}

func (p *Pipe) push() {
	n, err := io.Copy(p.left, p.right)
	p.down = n
	p.expire(err)
}

func (p *Pipe) expire(err error) {
	p.right.SetDeadline(time.Now())
	p.left.SetDeadline(time.Now())
	if err == nil {
		return
	}

	if err, ok := err.(net.Error); !ok || !err.Timeout() {
		p.err = err
	}
}

func (p *Pipe) close() {
	p.right.Close()
	logger.Debugf("pipe(%s) closing", p.PipeID)
}

func newPipe(l, r net.Conn) *Pipe {
	pid := fmt.Sprintf("%s<->%s", l.RemoteAddr().String(), r.RemoteAddr().String())
	p := &Pipe{
		PipeID: pid,
		left:   l,
		right:  r,
	}
	return p
}

type customer struct {
	sync.RWMutex
	address string
	aesKey  [32]byte
	pipes   map[string]*Pipe
}

func (cu *customer) addNewPipe(l net.Conn, target string) *Pipe {
	r, err := net.Dial("tcp", target)
	if err != nil {
		logger.Warningf("failed to connect target:->%v", err)
		return nil
	}
	r.(*net.TCPConn).SetKeepAlive(true)

	if debugConn == false {
		l, err = Shadow(l, cu.aesKey)
		if err != nil {
			logger.Warning("shadow the incoming conn err:->", err)
			return nil
		}
	}

	p := newPipe(l, r)
	cu.Lock()
	defer cu.Unlock()

	if _, ok := cu.pipes[p.PipeID]; ok {
		logger.Errorf("duplicate target:%s", p.PipeID)
		panic("this logic is wrong")
	}

	cu.pipes[p.PipeID] = p
	return p
}

func (cu *customer) removePipe(pid string) {

	cu.Lock()
	defer cu.Unlock()

	p, ok := cu.pipes[pid]
	if !ok {
		logger.Warning("no such pipe to remove:->", pid)
		return
	}
	p.close()
	delete(cu.pipes, pid)
	logger.Debugf("remove pipe(%s) from user(%s)", p.PipeID, cu.address)
}

func (cu *customer) isPipeEmpty() bool {
	return len(cu.pipes) == 0
}
