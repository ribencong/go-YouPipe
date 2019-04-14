package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
)

type HandShakeData struct {
	addr   string
	target string
}
type HandShake struct {
	ver string
	sig []byte
	*HandShakeData
}

func (s *HandShake) check() bool {
	msg, err := json.Marshal(s.HandShakeData)
	if err != nil {
		return false
	}

	pid, err := account.ConvertToID(s.addr)
	if err != nil {
		return false
	}
	return ed25519.Verify(pid.ToPubKey(), msg, s.sig)
}

func (node *SNode) handShake(conn *ctrlConn) {

	pipe, err := readServiceReq(conn, node)
	if err != nil {
		conn.writeAck(err)
		conn.Close()
		return
	}

	go pipe.pull()

	pipe.push()
}

func readServiceReq(conn *ctrlConn, node *SNode) (pipe *Pipe, err error) {
	req := &HandShake{}
	if err = conn.readMsg(req); err != nil {
		return nil, err
	}
	if req.check() {
		err = fmt.Errorf("signature invalid")
		return nil, err
	}

	cu := node.getCustomer(req.addr)
	if cu == nil {
		err = fmt.Errorf("micropayment channel isn't open")
		return nil, err
	}

	pipe, err = cu.pipeMng.addNewPipe(conn, req.target)
	if err != nil {
		return nil, err
	}

	return pipe, nil
}
