package client

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/service"
	"math/rand"
	"net"
	"strings"
)

const (
	MaxMinerSaved = 8
)

type Config struct {
	Addr        string
	Cipher      string
	LocalServer string
	License     string
	Services    []string
}

type AccountInfo struct {
	Addr   string
	Cipher string
}
type MinerInfo struct {
	minerAddr string
	minerIP   string
}

func (m MinerInfo) IsOK() bool {
	mid, err := account.ConvertToID(m.minerAddr)
	if err != nil {
		return false
	}
	port := mid.ToSocketPort()
	addr := network.JoinHostPort(m.minerIP, port) //TODO::set sole port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type YPServices []*MinerInfo

func (s YPServices) RandomService() *MinerInfo {

	r := rand.Intn(len(s))
	return s[r]
}

type Client struct {
	proxyServer net.Listener
	*account.Account
	connKey  service.PipeCryptKey
	license  *service.License
	services YPServices
	payCh    *PayChannel
}

func NewClient(conf *Config, password string) (*Client, error) {

	ls, err := net.Listen("tcp", conf.LocalServer)
	if err != nil {
		return nil, err
	}

	acc, err := account.AccFromString(conf.Addr, conf.Cipher, password)
	if err != nil {
		return nil, err
	}

	l, err := service.ParseLicense(conf.License)
	if err != nil {
		return nil, err
	}

	if l.UserAddr != acc.Address.ToString() {
		return nil, fmt.Errorf("license and account address are not same")
	}

	ser := PopulateService(conf.Services)
	if len(ser) == 0 {
		return nil, fmt.Errorf("no valid service")
	}

	c := &Client{
		Account:     acc,
		proxyServer: ls,
		license:     l,
		services:    ser,
	}

	if err := c.createPayChannel(); err != nil {
		return nil, err
	}

	go c.Proxying()

	return c, nil
}

func ParseService(path string) *MinerInfo {
	idIps := strings.Split(path, "@")

	if len(idIps) != 2 {
		fmt.Println("invalid path:", path)
		return nil
	}
	mi := &MinerInfo{
		minerAddr: idIps[0],
		minerIP:   idIps[1],
	}
	return mi
}

func PopulateService(paths []string) YPServices {
	s := make(YPServices, 0)

	var j = 0
	for _, path := range paths {
		mi := ParseService(path)
		if mi == nil || !mi.IsOK() {
			continue
		}

		s[j] = mi
		if j++; j >= MaxMinerSaved {
			break
		}
	}

	return s
}

func (c *Client) createPayChannel() error {
	port := c.Address.ToSocketPort() + 1 //TODO::

	mi := c.services.RandomService()

	addr := network.JoinHostPort(mi.minerIP, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	appConn := &service.CtrlConn{Conn: conn}

	data, err := json.Marshal(c.license)
	if err != nil {
		return err
	}

	if _, err := appConn.Write(data); err != nil {
		return err
	}

	ack := &service.ACK{}
	if err := appConn.ReadMsg(ack); err != nil {
		return err
	}

	c.payCh = &PayChannel{
		conn: appConn,
	}

	return nil
}
