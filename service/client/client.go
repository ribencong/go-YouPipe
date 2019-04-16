package client

import (
	"fmt"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/service"
	"math/rand"
	"net"
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

type Client struct {
	*account.Account
	proxyServer net.Listener
	aesKey      account.PipeCryptKey
	license     *service.License
	serverList  YPServices
	curService  *service.ServeNodeId
	payCh       *PayChannel
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

	ser := populateService(conf.Services)
	if len(ser) == 0 {
		return nil, fmt.Errorf("no valid service")
	}

	mi := ser.RandomService()

	c := &Client{
		Account:     acc,
		proxyServer: ls,
		license:     l,
		serverList:  ser,
		curService:  mi,
	}

	if err := account.GenerateAesKey(&c.aesKey,
		mi.ID.ToPubKey(), c.Key.PriKey); err != nil {
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
	addr := c.curService.ToPipeAddr()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	jsonConn := &service.JsonConn{Conn: conn}
	hs := &service.YPHandShake{
		CmdType:     service.CmdPayChanel,
		Sig:         c.license.Signature,
		LicenseData: c.license.LicenseData,
	}

	if err := jsonConn.Syn(hs); err != nil {
		return err
	}

	c.payCh = &PayChannel{
		conn:    jsonConn,
		done:    make(chan error),
		minerID: c.curService.ID,
		priKey:  c.Key.PriKey,
	}

	return nil
}

type YPServices []*service.ServeNodeId

func (s YPServices) RandomService() *service.ServeNodeId {
	r := rand.Intn(len(s))
	return s[r]
}

func populateService(paths []string) YPServices {
	s := make(YPServices, 0)

	var j = 0
	for _, path := range paths {
		mi := service.ParseService(path)
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
