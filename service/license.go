package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/utils"
	"golang.org/x/crypto/ed25519"
	"time"
)

const KingFinger = account.ID("YP5rttHPzRsAe2RmF52sLzbBk4jpoPwJLtABaMv6qn7kVm")

type JsonTime time.Time

func (jt JsonTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(utils.SysTimeFormat)+2)
	b = append(b, '"')
	b = time.Time(jt).AppendFormat(b, utils.SysTimeFormat)
	b = append(b, '"')
	return b, nil
}
func (jt *JsonTime) UnmarshalJSON(b []byte) error {
	t, err := time.ParseInLocation(`"`+utils.SysTimeFormat+`"`, string(b), time.Local)
	*jt = JsonTime(t)
	return err
}

func (jt JsonTime) String() string {
	return time.Time(jt).Format(utils.SysTimeFormat)
}

type LicenseData struct {
	StartDate JsonTime `json:"start"`
	EndDate   JsonTime `json:"end"`
	UserAddr  string   `json:"user"`
}

type License struct {
	Signature []byte `json:"sig"`
	*LicenseData
}

func (l *License) VerifySelf(sig []byte) error {

	data, err := json.Marshal(l)
	if err != nil {
		return err
	}

	selfID, err := account.ConvertToID(l.UserAddr)
	if err != nil {
		return err
	}

	if ok := ed25519.Verify(selfID.ToPubKey(), data, sig); !ok {
		return fmt.Errorf("signature verify self failed")
	}
	return nil
}

func (l *License) VerifyData() error {
	msg, err := json.Marshal(l.LicenseData)
	if err != nil {
		return err
	}

	if ok := ed25519.Verify(KingFinger.ToPubKey(), msg, l.Signature); !ok {
		return fmt.Errorf("signature verify data failed")
	}

	now := time.Now()
	if time.Time(l.EndDate).Before(now) || time.Time(l.StartDate).After(now) {
		return fmt.Errorf("lic time invalid(%s)", l.UserAddr)
	}

	return nil
}

func ParseLicense(data string) (*License, error) {
	l := &License{}
	if err := json.Unmarshal([]byte(data), l); err != nil {
		return nil, err
	}

	if err := l.VerifyData(); err != nil {
		return nil, err
	}
	return l, nil
}
