package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"time"
)

const KingFinger = account.ID("YP5rttHPzRsAe2RmF52sLzbBk4jpoPwJLtABaMv6qn7kVm")

type LicenseData struct {
	StartDate time.Time
	EndDate   time.Time
	UserAddr  string
}

type License struct {
	Signature []byte
	*LicenseData
}

func (l *License) Verify() error {
	msg, err := json.Marshal(l.LicenseData)
	if err != nil {
		return err
	}

	if ok := ed25519.Verify(KingFinger.ToPubKey(), msg, l.Signature); !ok {
		return fmt.Errorf("signature Verify failed")
	}

	now := time.Now()
	if now.Before(l.StartDate) || now.After(l.EndDate) {
		return fmt.Errorf("license time invalid(%s)", l.UserAddr)
	}

	return nil
}

func ParseLicense(data string) (*License, error) {
	l := &License{}
	if err := json.Unmarshal([]byte(data), l); err != nil {
		return nil, err
	}
	if err := l.Verify(); err != nil {
		return nil, err
	}
	return l, nil
}
