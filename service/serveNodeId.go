package service

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"net"
	"strings"
)

const ServeNodeSep = "@"

type ServeNodeId struct {
	ID account.ID
	IP string
}

func (m *ServeNodeId) IsOK() bool {
	addr := m.TONetAddr()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	hs := &YPHandShake{
		CmdType: CmdCheck,
	}

	jsonConn := JsonConn{conn}
	if err := jsonConn.Syn(hs); err != nil {
		return false
	}

	return true
}

func (m *ServeNodeId) TONetAddr() string {
	port := m.ID.ToServerPort()
	return network.JoinHostPort(m.IP, port)
}

func (m *ServeNodeId) ToString() string {
	return strings.Join([]string{m.ID.ToString(), m.IP}, ServeNodeSep)
}

func ParseService(path string) *ServeNodeId {
	idIps := strings.Split(path, ServeNodeSep)

	if len(idIps) != 2 {
		fmt.Println("invalid path:", path)
		return nil
	}

	id, err := account.ConvertToID(idIps[0])
	if err != nil {
		return nil
	}

	mi := &ServeNodeId{
		ID: id,
		IP: idIps[1],
	}
	return mi
}
