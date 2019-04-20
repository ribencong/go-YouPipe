package client

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/service"
	"golang.org/x/crypto/ed25519"
	"sync"
)

type PayChannel struct {
	sync.RWMutex
	minerID   account.ID
	priKey    ed25519.PrivateKey //TODO::
	conn      *service.JsonConn
	done      chan error
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

		fmt.Printf("Got new bill:%s", bill.String())

		proof, err := p.signBill(bill)
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

func (p *PayChannel) signBill(bill *service.PipeBill) (*service.PipeProof, error) {

	if ok := bill.Verify(p.minerID); !ok {
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

	if err := proof.Sign(p.priKey); err != nil {
		return nil, err
	}

	p.unSigned -= bill.UsedBandWidth
	p.totalUsed += bill.UsedBandWidth
	return proof, nil
}

func (p *PayChannel) Consume(n int) {
	p.Lock()
	defer p.Unlock()
	p.unSigned += int64(n)
}
