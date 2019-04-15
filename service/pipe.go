package service

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"io"
	"net"
	"sync"
	"time"
)

type PipeAdmin struct {
	sync.RWMutex
	aesKey PipeCryptKey
	pipes  map[string]*Pipe
}

func (a *PipeAdmin) IsEmpty() bool {
	a.RLock()
	defer a.RUnlock()
	return len(a.pipes) == 0
}

func newAdmin(peerAddr string) *PipeAdmin {
	admin := &PipeAdmin{
		pipes: make(map[string]*Pipe),
	}
	if err := account.GetAccount().CreateAesKey((*[32]byte)(&admin.aesKey), peerAddr); err != nil {
		logger.Errorf("create pipe admin's aes key err:->", err)
		return nil
	}

	return admin
}

func (a *PipeAdmin) addNewPipe(l net.Conn, target string) (*Pipe, error) {
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

type Pipe struct {
	PipeID string
	err    error
	up     int64
	down   int64
	left   net.Conn
	right  net.Conn
}

func (p *Pipe) pullFromServer() {
	n, err := io.Copy(p.right, p.left)
	p.up = n
	p.expire(err)
}

func (p *Pipe) pushBackToClient() {
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
