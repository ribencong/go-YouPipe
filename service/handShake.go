package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
)

type HandShakeData struct {
	Addr   string
	Target string
}
type HandShake struct {
	Ver string
	Sig []byte
	*HandShakeData
}

func (s *HandShake) check() bool {
	msg, err := json.Marshal(s.HandShakeData)
	if err != nil {
		return false
	}

	pid, err := account.ConvertToID(s.Addr)
	if err != nil {
		return false
	}
	return ed25519.Verify(pid.ToPubKey(), msg, s.Sig)
}

func (node *PipeMiner) handShake(conn *CtrlConn) {

	pipe, err := readServiceReq(conn, node)
	if err != nil {
		return
	}

	go pipe.pullFromServer()

	pipe.pushBackToClient()
}

func readServiceReq(conn *CtrlConn, node *PipeMiner) (pipe *Pipe, err error) {
	req := &HandShake{}
	if err = conn.ReadMsg(req); err != nil {
		return nil, err
	}

	var cu *service = nil
	if req.check() {
		err = fmt.Errorf("signature invalid")
		goto ACK
	}

	cu = node.getCustomer(req.Addr)
	if cu == nil {
		err = fmt.Errorf("micropayment channel isn't open")
		goto ACK
	}

	pipe, err = cu.pipeMng.addNewPipe(conn, req.Target)
	if err != nil {
		goto ACK
	}

ACK:
	conn.writeAck(err)
	return
}
