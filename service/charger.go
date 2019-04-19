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
	done       chan error
	token      int64
	peerID     account.ID
	bill       chan *PipeBill
	receipt    chan *PipeProof
	checkIn    chan struct{}
	peerIPAddr string
	aesKey     account.PipeCryptKey
}

func (c *bandCharger) waitingReceipt() {
	defer logger.Infof("waiting receipt thread exit:%s", c.peerID)

	for {
		proof := &PipeProof{}
		if err := c.ReadJsonMsg(proof); err != nil {
			logger.Warning("read bandwidth proof err:->", err)
			c.done <- err
			return
		}

		if !proof.Verify(c.peerID) {
			logger.Error("wrong signature for bandwidth bill:->", proof)
			continue
		}

		c.fullFill(proof.UsedBandWidth)
		c.receipt <- proof
	}
}

func (c *bandCharger) charging() {
	defer logger.Infof("charging thread exit:%s", c.peerID)
	defer c.Close()

	for {
		select {
		case bill := <-c.bill:
			if err := c.WriteJsonMsg(bill); err != nil {
				logger.Error("charger channel err:->", err)
				c.done <- err
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *bandCharger) fullFill(used int64) {
	c.Lock()
	defer c.Unlock()
	logger.Debugf("microPay from :%s with :%d and token is:%d", c.peerID.ToString(), used, c.token)
	c.token += used
	c.checkIn <- struct{}{}
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
	case <-c.checkIn:
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
		Ver:           CurrentMineralVer,
		MinedTime:     time.Now(),
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
