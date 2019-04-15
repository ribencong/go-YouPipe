package client

import (
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"math/rand"
	"net"
)

type MinerInfo struct {
	minerAddr account.ID
	minerIP   string
}

func (m MinerInfo) IsOK() bool {

	port := m.minerAddr.ToServerPort()
	addr := network.JoinHostPort(m.minerIP, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (m MinerInfo) ToPipeAddr() string {
	port := m.minerAddr.ToServerPort()
	return network.JoinHostPort(m.minerIP, port)
}

type YPServices []*MinerInfo

func (s YPServices) RandomService() *MinerInfo {
	r := rand.Intn(len(s))
	return s[r]
}
