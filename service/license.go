package service

import (
	"encoding/json"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"time"
)

const KingFinger = account.ID("YP5rttHPzRsAe2RmF52sLzbBk4jpoPwJLtABaMv6qn7kVm")

type LicenseContent struct {
	StartDate time.Time
	EndDate   time.Time
	UserAddr  string
}

type License struct {
	Signature []byte
	Content   []byte
}

type LicenseCheckResult struct {
	Success bool
	ErrMsg  string
}

func CheckLicense(l *License) bool {

	msg, err := json.Marshal(l.Content)
	if err != nil {
		logger.Warning(err)
		return false
	}

	return ed25519.Verify(KingFinger.ToPubKey(), msg, l.Signature)
}
