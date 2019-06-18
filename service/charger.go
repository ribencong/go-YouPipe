package service

import (
	"encoding/json"
	"fmt"
	"github.com/ribencong/go-youPipe/account"
	"sync"
	"time"
)

const (
	BandWidthPerToPay = 1 << 24 //
	BillThreshold     = 1 << 22 //
	MaxBandBill       = 4       //
)

//TODO::Add keep alive logic
type bandCharger struct {
	*JsonConn
	sync.RWMutex
	done       chan error
	token      int64
	used       int64
	peerID     account.ID
	billID     int
	bill       chan *PipeBill
	receipt    chan *PipeProof
	peerIPAddr string
	aesKey     account.PipeCryptKey
}

//TODO:: Add keep alive message logic.
func (c *bandCharger) waitingReceipt() {
	defer logger.Infof("waiting receipt thread exit:%s", c.peerID)

	for {
		proof := &PipeProof{}
		if err := c.ReadJsonMsg(proof); err != nil {
			logger.Warningf("(%s)read bandwidth proof err:%s->", c.peerID, err.Error())
			c.done <- err
			return
		}

		if !proof.Verify(c.peerID) {
			err := fmt.Errorf("(%s)wrong signature for bandwidth bill:%v", c.peerID, proof)
			logger.Warning(err)
			c.done <- err
			return
		}

		go func() {
			c.fullFill(proof.UsedBandWidth)
			c.receipt <- proof
			logger.Notice(proof)
		}()

	}
}

func (c *bandCharger) charging() {
	defer logger.Infof("charging thread exit:%s", c.peerID)
	for {
		select {
		case bill := <-c.bill:
			if err := c.WriteJsonMsg(bill); err != nil {
				logger.Warningf("charger(%s) channel err:%s->", c.peerID, err.Error())
				c.done <- err
				return
			}
		case <-c.done:
			logger.Warningf("(%s)charger received done signal", c.peerID)
			return
		}
	}
}

func (c *bandCharger) fullFill(used int64) {

	c.Lock()
	defer c.Unlock()
	c.token += used
	logger.Noticef("microPay from :%s with :%d and token is:%d", c.peerID.ToString(), used, c.token)
}

func (c *bandCharger) Charge(n int) error {
	c.Lock()
	defer c.Unlock()

	c.token -= int64(n)
	c.used += int64(n)

	logger.Infof("(%s)charged:token:%d, used:%d, toSub:%d",
		c.peerID, c.token, c.used, n)

	if c.used >= BillThreshold*2 {
		c.billID++
		c.bill <- createBill(c.peerID.ToString(), BillThreshold, c.billID)
		c.used -= BillThreshold
	}

	if c.token <= 0 {
		logger.Warningf("bill for (%s) time out", c.peerID)
		return fmt.Errorf("time out")
	}

	return nil
}

func (c *bandCharger) finish() {
	logger.Noticef("charger(%s) finished", c.peerID)
	c.Close()
}

func createBill(customerAddr string, usedBand int64, id int) *PipeBill {

	mi := &Mineral{
		ID:            id,
		MinedTime:     time.Now(),
		UsedBandWidth: usedBand,
		ConsumerAddr:  customerAddr,
		MinerAddr:     account.GetAccount().Address.ToString(),
	}

	data, _ := json.Marshal(mi)
	sig := account.GetAccount().Sign(data)
	logger.Noticef("New bill:%s", data)

	return &PipeBill{
		MinerSig: sig,
		Mineral:  mi,
	}
}
