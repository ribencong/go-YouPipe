package thread

import (
	"context"
	"fmt"
	"github.com/youpipe/go-youPipe/utils"
	"sync"
	"time"
)

type Thread struct {
	sync.RWMutex
	ID   int
	Name string
	IsOk bool
	ctx  context.Context
	exit context.CancelFunc
	r    RunnerI
	c    time.Time
}

type RunnerI interface {
	Run(ctx context.Context)
	CloseCallBack(t *Thread)
	DebugInfo() string
}

func NewThreadWithName(ru RunnerI, name string) *Thread {
	t := NewThread(ru)
	t.Name = name
	return t
}

func NewThread(ru RunnerI) *Thread {

	c, s := context.WithCancel(context.Background())
	threadCounter.Lock()
	defer threadCounter.Unlock()

	threadCounter.id++
	id := threadCounter.id

	t := &Thread{
		ID:   id,
		ctx:  c,
		IsOk: false,
		exit: s,
		r:    ru,
		c:    time.Now(),
	}

	threadCounter.cache[id] = t
	return t
}

func (t *Thread) Stop() {
	t.Lock()
	defer t.Unlock()

	t.exit()
}

func (t *Thread) Start() {
	t.RLock()
	defer t.RUnlock()

	if t.IsOk {
		logger.Info("using a running thread")
		return
	}

	go func() {

		defer clearCache(t.ID)

		t.Lock()
		t.IsOk = true
		t.Unlock()

		logger.Infof("Thread (%d, %s) start", t.ID, t.Name)
		t.r.Run(t.ctx)
		logger.Infof("Thread (%d, %s) exit", t.ID, t.Name)

		t.Lock()
		t.IsOk = false
		t.r.CloseCallBack(t)
		t.Unlock()
	}()
}

func (t *Thread) ThreadDebugInfo() string {
	t.RLock()
	defer t.RUnlock()

	str := fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++\n"+
		"+%-15s:%30d+\n"+
		"+%-15s:%30t+\n"+
		"+%-15s:%30s+\n"+
		"%s\n"+
		"\n++++++++++++++++++++++++++++++++++++++++++++++++",
		"ID", t.ID,
		"status", t.IsOk,
		"createTime", t.c.Format(utils.SysTimeFormat),
		t.r.DebugInfo())

	return str
}
