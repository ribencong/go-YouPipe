package service

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"net"
	"strings"
	"time"
)

const ServeNodeSep = "@"
const ServeNodeTimeOut = time.Second * 2

type ServeNodeId struct {
	ID   account.ID
	IP   string
	Ping time.Duration
}

func (m *ServeNodeId) IsOK() bool {

	addr := m.TONetAddr()
	conn, err := net.DialTimeout("tcp", addr, ServeNodeTimeOut)
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

func IsIPAddr(ip string) bool {
	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		fmt.Printf("%v is not a valid IPv4 address\n", trial)

		if trial.To16() == nil {
			fmt.Printf("%v is not a valid IP address\n", trial)
			return false
		}
	}

	return true
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

	if ok := IsIPAddr(idIps[1]); !ok {
		return nil
	}

	mi := &ServeNodeId{
		ID:   id,
		IP:   idIps[1],
		Ping: time.Hour, //Default is big value
	}
	return mi
}
