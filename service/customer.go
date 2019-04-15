package service

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/thread"
	"golang.org/x/crypto/ed25519"
	"net"
)

type customer struct {
	peerID     account.ID
	done       chan error
	aesKey     PipeCryptKey
	pipes      map[string]*RightPipe
	license    *License
	payChannel *BWCharger
}

func (node *PipeMiner) newCustomer(c net.Conn) {
	defer c.Close()

	jsonConn := &JsonConn{c}

	l := &License{}
	if err := jsonConn.ReadJsonMsg(l); err != nil {
		logger.Errorf("read license of customer err:%v", err)
		return
	}

	ack := &YouPipeACK{Success: false}
	var key PipeCryptKey

	if err := l.Verify(); err != nil {
		ack.Message = fmt.Sprintf("read license of customer err:%v", err)
		goto ACK
	}

	if cu := node.getCustomer(l.UserAddr); cu != nil {
		ack.Message = fmt.Sprint("Duplicate customer online")
		goto ACK
	}

	peerID := account.ID(l.UserAddr)
	if err := account.GenerateAesKey((*[32]byte)(&key),
		peerID.ToPubKey(), c.Key.PriKey); err != nil {
		return nil, err
	}

	ack.Success = true
ACK:
	jsonConn.WriteJsonMsg(ack)
	if ack.Success {
		return
	}

	user := &customer{
		peerID:  peerID,
		license: l,
		done:    make(chan error),
		payChannel: &BWCharger{
			JsonConn: jsonConn,
			priKey:   node.Key.PriKey,
		},
	}

	node.addCustomer(l.UserAddr, user)

	user.working()
}

func (c *customer) working() {
	c.done = make(chan struct{})
}

func (c *customer) destroy() {
}
