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

func NewHandReq(addr, target string, key ed25519.PrivateKey) *HandShake {
	reqData := &HandShakeData{
		Addr:   addr,
		Target: target,
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("marshal hand shake data err:->", err)
		return nil
	}

	sig := ed25519.Sign(key, data)
	req := &HandShake{
		Sig:           sig,
		HandShakeData: reqData,
	}

	return req
}

func (node *SNode) handShake(conn *CtrlConn) {

	pipe, err := readServiceReq(conn, node)
	if err != nil {
		conn.writeAck(err)
		conn.Close()
		return
	}

	go pipe.pull()

	pipe.push()
}

func readServiceReq(conn *CtrlConn, node *SNode) (pipe *Pipe, err error) {
	req := &HandShake{}
	if err = conn.ReadMsg(req); err != nil {
		return nil, err
	}
	if req.check() {
		err = fmt.Errorf("signature invalid")
		return nil, err
	}

	cu := node.getCustomer(req.Addr)
	if cu == nil {
		err = fmt.Errorf("micropayment channel isn't open")
		return nil, err
	}

	pipe, err = cu.pipeMng.addNewPipe(conn, req.Target)
	if err != nil {
		return nil, err
	}

	return pipe, nil
}
