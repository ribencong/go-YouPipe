package service

import (
	"encoding/json"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
)

type Mineral struct {
	UsedBandWidth int64
	ConsumerAddr  string
	MinerAddr     string
}

type PipeBill struct {
	MinerSig []byte
	*Mineral
}

func (b PipeBill) Sign(priKey ed25519.PrivateKey) error {
	data, err := json.Marshal(b.Mineral)
	if err != nil {
		return err
	}

	b.MinerSig = ed25519.Sign(priKey, data)
	return nil
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

func (p PipeProof) Sign(priKey ed25519.PrivateKey) error {
	data, err := json.Marshal(p.PipeBill)
	if err != nil {
		return err
	}

	p.ConsumerSig = ed25519.Sign(priKey, data)
	return nil
}

func (p PipeProof) Verify(addr account.ID) bool {
	data, err := json.Marshal(p.PipeBill)
	if err != nil {
		return false
	}

	return ed25519.Verify(addr.ToPubKey(), data, p.ConsumerSig)
}
