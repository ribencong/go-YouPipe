package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"sync"
	"time"
)

const (
	BandWidthPerToPay = 1 << 22 //4M
	SignBillTimeOut   = time.Second * 4
)

type bandCharger struct {
	*JsonConn
	sync.RWMutex
	token      int64
	peerID     account.ID
	bill       chan *PipeBill
	receipt    chan struct{}
	peerIPAddr string
	aesKey     account.PipeCryptKey
}

func (c *bandCharger) charging() error {

	for {
		bill := <-c.bill
		if err := c.WriteJsonMsg(bill); err != nil {
			logger.Error("charger channel err:->", err)
			return err
		}

		proof := &PipeProof{}
		if err := c.ReadJsonMsg(proof); err != nil {
			logger.Error("read bandwidth proof err:->", err)
			return err
		}

		if proof.Verify(c.peerID) {
			c.fullFill()
		} else {
			logger.Error("wrong signature for bandwidth bill:->", proof)
		}
	}
}

func (c *bandCharger) fullFill() {
	c.Lock()
	defer c.Unlock()
	c.token += BandWidthPerToPay
	c.receipt <- struct{}{}
}

func (c *bandCharger) Charge(n int) error {
	c.Lock()
	defer c.Unlock()
	c.token -= int64(n)

	if c.token > BandWidthPerToPay {
		return nil
	}

	c.bill <- createBill(c.peerID.ToString())
	select {
	case <-c.receipt:
		{
			return nil
		}
	case <-time.After(SignBillTimeOut):
		{
			return fmt.Errorf("time out")
		}
	}

	return nil
}

func createBill(customerAddr string) *PipeBill {

	mi := &Mineral{
		UsedBandWidth: BandWidthPerToPay,
		ConsumerAddr:  customerAddr,
		MinerAddr:     account.GetAccount().Address.ToString(),
	}

	data, _ := json.Marshal(mi)
	sig := account.GetAccount().Sign(data)

	return &PipeBill{
		MinerSig: sig,
		Mineral:  mi,
	}
}
