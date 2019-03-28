package gossip

import (
	"github.com/youpipe/go-youPipe/thread"
	"time"
)

type TimerTask map[int]*thread.TimerThread

func (t TimerTask) Add(du time.Duration, ru thread.TimerThreadRun, name string) *thread.TimerThread {
	tt := thread.NewTimerThread(du, ru)
	t[tt.ID] = tt
	tt.Name = name
	return tt
}

func (t TimerTask) StartAll() {
	for _, task := range t {
		task.Run()
	}
}

func (t TimerTask) RemoveAll() {
	for _, task := range t {
		task.Close()
	}
}

func (n *GNode) isolateCheck() {
	if !n.income.IsEmpty() {
		return
	}

	logger.Warning("no input view nodes")

	if n.outPut.IsEmpty() {
		logger.Warning("I'm stand alone.....")
		n.subscribe()
		return
	}

	data, _ := pack(ReSubscribe, n.NodeID)
	if err := n.outPut.RandomSend(data); err != nil {
		logger.Warning("error when resubscribe:->", err)
	}
}

func (n *GNode) heartBeatTimer() {
	logger.Debug("sending heart beat......")

	data, _ := pack(HeartBeat, n.NodeID)

	n.outPut.Broadcast(data)

	if !n.expired.IsZero() && n.expired.Before(time.Now()) {

		logger.Info("expired time: clock stop and resubscribe")
		n.income.ClearAll()

		n.expired = time.Time{}
	}
}
