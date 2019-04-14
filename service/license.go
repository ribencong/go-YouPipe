package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"time"
)

const KingFinger = account.ID("YP5rttHPzRsAe2RmF52sLzbBk4jpoPwJLtABaMv6qn7kVm")

type Content struct {
	StartDate time.Time
	EndDate   time.Time
	UserAddr  string
}

type License struct {
	Signature []byte
	*Content
}

func (l *License) check() bool {
	msg, err := json.Marshal(l.Content)
	if err != nil {
		return false
	}

	if ok := ed25519.Verify(KingFinger.ToPubKey(), msg, l.Signature); !ok {
		logger.Warning("signature check failed")
		return false
	}
	now := time.Now()
	if now.Before(l.StartDate) || now.After(l.EndDate) {
		logger.Warning("license time invalid(%s)", l.UserAddr)
		return false
	}

	return false
}

func initCustomer(conn *CtrlConn, node *SNode) (*customer, error) {
	l := &License{}
	if err := conn.ReadMsg(l); err != nil {
		return nil, err
	}

	if !l.check() {
		return nil, fmt.Errorf("signature failed err:%s", l.UserAddr)
	}

	peerAddr := l.UserAddr
	cu := node.getCustomer(peerAddr)

	if cu == nil {
		admin := newAdmin(peerAddr)
		if admin == nil {
			return nil, fmt.Errorf("aes key error when create customer%s", peerAddr)
		}

		cu = &customer{
			address:    peerAddr,
			license:    l,
			pipeMng:    admin,
			payChannel: newMicroPayment(conn),
		}

		node.addCustomer(peerAddr, cu)
	}

	conn.writeAck(nil)
	return cu, nil
}
