package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"time"
)

const (
	CurrentMineralVer = 1
)

type Mineral struct {
	Ver           int
	MinedTime     time.Time
	UsedBandWidth int64
	ConsumerAddr  string
	MinerAddr     string
}

type PipeBill struct {
	MinerSig       []byte
	ClientSignTime time.Time
	*Mineral
}

func (b PipeBill) Verify(addr account.ID) bool {
	data, err := json.Marshal(b.Mineral)
	if err != nil {
		return false
	}

	return ed25519.Verify(addr.ToPubKey(), data, b.MinerSig)
}

type PipeProof struct {
	*PipeBill
	ConsumerSig []byte
}

func (p *PipeProof) Sign(priKey ed25519.PrivateKey) error {
	p.ClientSignTime = time.Now()

	data, err := json.Marshal(p.PipeBill)
	if err != nil {
		return err
	}

	p.ConsumerSig = ed25519.Sign(priKey, data)
	return nil
}

func (p *PipeProof) Verify(addr account.ID) bool {
	data, err := json.Marshal(p.PipeBill)
	if err != nil {
		return false
	}

	return ed25519.Verify(addr.ToPubKey(), data, p.ConsumerSig)
}

func (p *PipeProof) ToData() []byte {
	data, _ := json.Marshal(p)
	return data
}

func (p *PipeProof) ToID() string {
	return fmt.Sprintf("%s@%s", p.MinerSig, p.ConsumerSig)
}
