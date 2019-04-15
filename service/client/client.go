package client

import (
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
	minerAddr account.ID
	minerIP   string
}

func (m MinerInfo) IsOK() bool {

	port := m.minerAddr.ToSocketPort()
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
	*account.Account
	proxyServer     net.Listener
	connKey         service.PipeCryptKey
	license         *service.License
	services        YPServices
	selectedService *MinerInfo
	payCh           *PayChannel
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

	mi := ser.RandomService()

	c := &Client{
		Account:         acc,
		proxyServer:     ls,
		license:         l,
		services:        ser,
		selectedService: mi,
	}

	if err := account.GenerateAesKey((*[32]byte)(&c.connKey),
		mi.minerAddr.ToPubKey(), c.Key.PriKey); err != nil {
		return nil, err
	}

	if err := c.createPayChannel(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Running() error {

	go c.payCh.payMonitor()

	go c.Proxying()

	for {
		select {
		case err := <-c.payCh.done:
			return err
		}
	}
}
func (c *Client) createPayChannel() error {
	port := c.Address.ToSocketPort() + 1 //TODO::

	addr := network.JoinHostPort(c.selectedService.minerIP, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	appConn := &service.JsonConn{Conn: conn}

	if err := appConn.WriteJsonMsg(c.license); err != nil {
		return err
	}

	ack := &service.ACK{}
	if err := appConn.ReadJsonMsg(ack); err != nil {
		return err
	}

	if !ack.Success {
		return fmt.Errorf("create payment channel failed:%s", ack.Message)
	}

	ch := &PayChannel{
		conn:   appConn,
		done:   make(chan error),
		peerID: c.Address,
		PriKey: c.Key.PriKey,
	}

	c.payCh = ch
	return nil
}

func ParseService(path string) *MinerInfo {
	idIps := strings.Split(path, "@")

	if len(idIps) != 2 {
		fmt.Println("invalid path:", path)
		return nil
	}

	id, err := account.ConvertToID(idIps[0])
	if err != nil {
		return nil
	}

	mi := &MinerInfo{
		minerAddr: id,
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
