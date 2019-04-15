package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"net"
)

type PipeCmd int

const (
	_ PipeCmd = iota
	CmdPayChanel
	CmdPipe
	CmdCheck
)

type YouPipeACK struct {
	Success bool
	Message string
}

type YPHandShake struct {
	CmdType PipeCmd
	Sig     []byte
	*PipeReqData
	*LicenseData
}

func (s *PipeRequest) check() bool {
	msg, err := json.Marshal(s.PipeReqData)
	if err != nil {
		return false
	}

	pid, err := account.ConvertToID(s.Addr)
	if err != nil {
		return false
	}
	return ed25519.Verify(pid.ToPubKey(), msg, s.Sig)
}

func readServiceReq(conn *JsonConn, node *PipeMiner) (pipe *RightPipe, err error) {

	var cu *customer = nil
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
