package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"net"
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

	return ed25519.Verify(KingFinger.ToPubKey(), msg, l.Signature)
}

func getValidLicense(conn net.Conn, node *SNode) (*License, error) {
	var err error
	defer conn.Write(sealACK(err))

	buff := make([]byte, buffSize)
	n, err := conn.Read(buff)
	if err != nil {
		err = fmt.Errorf("payment channel open err:%v", err)
		return nil, err
	}

	l := &License{}
	if err = json.Unmarshal(buff[:n], l); err != nil {
		err = fmt.Errorf("payment channel read data err:%v", err)
		return nil, err
	}

	peerAddr := l.UserAddr
	if !l.check() {
		err = fmt.Errorf("signature failed err")
		return nil, err
	}

	now := time.Now()
	if now.Before(l.StartDate) || now.After(l.EndDate) {
		err = fmt.Errorf("license time invalid(%s)", peerAddr)
		return nil, err
	}

	if u := node.getCustomer(peerAddr); u != nil {
		err = fmt.Errorf("duplicate customer server (%s)", peerAddr)
		return nil, err
	}

	return l, nil
}
