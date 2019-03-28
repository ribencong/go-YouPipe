package thread

import (
	"context"
	"fmt"
	"github.com/youpipe/go-node/utils"
	"sync"
	"time"
)

type TimerThreadRun func()

var threadCounter = struct {
	id    int
	cache map[int]interface{}
	sync.RWMutex
}{
	id:    1,
	cache: make(map[int]interface{}),
}
var logger = utils.NewLog(utils.LMThread)

type TimerThread struct {
	sync.RWMutex
	ID    int
	Name  string
	IsOk  bool
	ctx   context.Context
	Close context.CancelFunc
	d     time.Duration
	r     TimerThreadRun
	c     time.Time
}

func NewTimerThread(du time.Duration, ru TimerThreadRun) *TimerThread {
	c, s := context.WithCancel(context.Background())

	threadCounter.Lock()
	defer threadCounter.Unlock()

	threadCounter.id++
	id := threadCounter.id

	tt := &TimerThread{
		ID:    id,
		ctx:   c,
		IsOk:  false,
		Close: s,
		d:     du,
		r:     ru,
		c:     time.Now(),
	}
	threadCounter.cache[id] = tt
	return tt
}

func clearCache(id int) {
	threadCounter.Lock()
	defer threadCounter.Unlock()
	delete(threadCounter.cache, id)
}

func ThreadNO() int {
	threadCounter.RLock()
	defer threadCounter.RUnlock()
	return len(threadCounter.cache)
}
func ShowThreadInfos() string {

	threadCounter.RLock()
	defer threadCounter.RUnlock()

	str := fmt.Sprintf("\n\n*************************(%d)*************************",
		len(threadCounter.cache))

	for _, t := range threadCounter.cache {
		if tt, ok := t.(*TimerThread); ok {
			str += tt.ThreadDebugInfo()
		} else {
			str += t.(*Thread).ThreadDebugInfo()
		}
	}

	str += fmt.Sprintf("\n\n********************************************************")
	return str
}

func (t *TimerThread) Run() {
	t.RLock()
	defer t.RUnlock()
	if t.IsOk {
		logger.Warning("duplicate time thread run")
		return
	}

	go func() {
		t.Lock()
		t.IsOk = true
		t.Unlock()

		logger.Debugf("TimerThread (%d<=>%s) start", t.ID, t.Name)
		defer clearCache(t.ID)
		for {
			select {
			case <-time.After(t.d):
				t.r()
			case <-t.ctx.Done():
				logger.Debugf("TimerThread (%d<=>%s) closed", t.ID, t.Name)
				t.Lock()
				t.IsOk = false
				t.Unlock()
				return
			}
		}
	}()
}
func (t *TimerThread) ThreadName() string {
	return t.Name
}
func (t *TimerThread) ThreadDebugInfo() string {
	t.RLock()
	defer t.RUnlock()

	str := fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++\n"+
		"+%-15s:%20d+\n"+
		"+%-15s:%20s+\n"+
		"+%-15s:%20t+\n"+
		"+%-15s:%20s+\n"+
		"+%-15s:%20.2f+\n"+
		"++++++++++++++++++++++++++++++++++++++",
		"ID", t.ID,
		"name", t.Name,
		"status", t.IsOk,
		"createTime", t.c.Format(utils.SysTimeFormat),
		"duration", t.d.Seconds())

	return str
}
