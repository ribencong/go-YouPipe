package service

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"net"
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

func (node *SNode) handShake(conn net.Conn) {

	pipe, err := readServiceReq(conn, node)
	if err != nil {
		conn.Close()
		return
	}
	go pipe.pull()
	pipe.push()
}

func readServiceReq(conn net.Conn, node *SNode) (*Pipe, error) {
	var err error
	defer conn.Write(sealACK(err))

	buffer := make([]byte, buffSize)
	n, err := conn.Read(buffer)
	if err != nil {
		err = fmt.Errorf("failed to read address:->%v", err)
		return nil, err
	}

	sockReq := &HandShake{}
	if err = json.Unmarshal(buffer[:n], sockReq); err != nil {
		err = fmt.Errorf("unmarshal address:->%v", err)
		return nil, err
	}

	if sockReq.check() {
		err = fmt.Errorf("signature invalid")
		return nil, err
	}

	cu := node.getCustomer(sockReq.addr)
	if cu == nil {
		err = fmt.Errorf("micropayment channel isn't open")
		return nil, err
	}

	pipe, err := cu.pipeMng.addNewPipe(conn, sockReq.target)
	if err != nil {
		return nil, err
	}

	return pipe, nil
}
