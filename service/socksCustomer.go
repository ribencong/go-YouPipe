package service

import (
	"io"
	"net"
	"sync"
	"time"
)

type Pipe struct {
	err   error
	up    int64
	down  int64
	left  net.Conn
	right net.Conn
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
}

type customer struct {
	sync.RWMutex
	address string
	aesKey  [32]byte
	pipes   map[string]*Pipe
}

func (cu *customer) addNewPipe(l net.Conn, target string, raw bool) *Pipe {
	r, err := net.Dial("tcp", target)
	if err != nil {
		logger.Warningf("failed to connect target:->%v", err)
		return nil
	}
	r.(*net.TCPConn).SetKeepAlive(true)

	if raw == false {
		l, err = Shadow(l, cu.aesKey)
		if err != nil {
			logger.Warning("shadow the incoming conn err:->", err)
			return nil
		}
	}

	p := &Pipe{
		left:  l,
		right: r,
	}

	cu.Lock()
	defer cu.Unlock()

	if _, ok := cu.pipes[target]; ok {
		panic("this logic is wrong")
	}

	cu.pipes[target] = p
	return p
}

func (cu *customer) removePipe(target string) {

	cu.Lock()
	defer cu.Unlock()

	u, ok := cu.pipes[target]
	if !ok {
		logger.Warning("no such pipe to remove:->", target)
		return
	}
	u.close()
	delete(cu.pipes, target)
}

func (cu *customer) isPipeEmpty() bool {
	return len(cu.pipes) == 0
}
