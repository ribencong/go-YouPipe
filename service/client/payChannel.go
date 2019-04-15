package client

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/service"
	"sync"
)

type PayChannel struct {
	*Client
	conn *service.JsonConn
	done chan error
	sync.RWMutex
	totalUsed int64
	unSigned  int64
}

func (p *PayChannel) payMonitor() {

	for {

		bill := &service.PipeBill{}
		if err := p.conn.ReadJsonMsg(bill); err != nil {
			p.done <- fmt.Errorf("payment channel closed: %v", err)
			return
		}

		proof, err := p.Sign(bill)
		if err != nil {
			p.done <- err
			return
		}

		if err := p.conn.WriteJsonMsg(proof); err != nil {
			p.done <- err
			return
		}
	}
}

func (p *PayChannel) Sign(bill *service.PipeBill) (*service.PipeProof, error) {

	minerId, _ := account.ConvertToID(p.selectedService.minerAddr)
	if ok := bill.Verify(minerId); ok {
		return nil, fmt.Errorf("miner's signature failed")
	}

	p.Lock()
	defer p.Unlock()

	if bill.UsedBandWidth > p.unSigned {
		return nil, fmt.Errorf("I don't use so much bandwith user:(%d) unsigned(%d)", bill.UsedBandWidth, p.unSigned)
	}

	proof := &service.PipeProof{
		PipeBill: bill,
	}

	if err := proof.Sign(p.Account.Key.PriKey); err != nil {
		return nil, err
	}

	p.unSigned -= bill.UsedBandWidth
	p.totalUsed += bill.UsedBandWidth
	return proof, nil
}
