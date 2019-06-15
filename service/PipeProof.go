package service

import (
	"encoding/json"
	"fmt"
	"github.com/ribencong/go-youPipe/account"
	"github.com/ribencong/go-youPipe/utils"
	"golang.org/x/crypto/ed25519"
	"time"
)

const (
	CurrentMineralVer = 1
)

type Mineral struct {
	ID            int
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

func (b *PipeBill) String() string {
	return fmt.Sprintf("ID=%d, MinedTime=%s, UsedBandWidth=%d,"+
		"ConsumerAddr=%s, MinerAddr=%s",
		b.ID, b.MinedTime.Format(utils.SysTimeFormat),
		b.UsedBandWidth, b.ConsumerAddr,
		b.MinerAddr)
}

func (b *PipeBill) Verify(pubKey ed25519.PublicKey) bool {
	data, err := json.Marshal(b.Mineral)
	if err != nil {
		return false
	}

	return ed25519.Verify(pubKey, data, b.MinerSig)
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
