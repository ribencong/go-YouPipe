package client

import (
	"encoding/json"
	"fmt"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/service"
	"golang.org/x/crypto/ed25519"
	"net"
)

func (c *Client) Proxying() {
	conn, err := c.proxyServer.Accept()
	if err != nil {
		fmt.Printf("finish to accept :%s", err)
		return
	}
	go c.consume(conn)

}

func (c *Client) consume(conn net.Conn) {
	defer conn.Close()
	conn.(*net.TCPConn).SetKeepAlive(true)
	fmt.Println("a new connection :->", conn.RemoteAddr().String())

	obj, err := HandShake(conn)
	if err != nil {
		fmt.Println("sock5 handshake err:->", err)
		return
	}
	fmt.Println("target info:->", obj.target)

	port := c.selectedService.minerAddr.ToSocketPort()
	addr := network.JoinHostPort(c.selectedService.minerIP, port)
	rConn, err := net.Dial("tcp", addr)

	if err != nil {
		fmt.Printf("failed to connect to (%s) access point server (%s):->", addr, err)
		return
	}
	rConn.(*net.TCPConn).SetKeepAlive(true)

	consumeConn := service.NewConsumerConn(rConn, c.connKey)
	if consumeConn == nil {
		fmt.Println("create consume Conn failed")
		return
	}
	req := c.NewHandReq(obj.target)

	if err := consumeConn.WriteJsonMsg(req); err != nil {
		fmt.Println("write hand shake data err:->", err)
		return
	}
	ack := &service.ACK{}
	if err := consumeConn.ReadJsonMsg(ack); err != nil {
		fmt.Printf("failed to read miner's response :->%v", err)
		return
	}

	if !ack.Success {
		fmt.Println("hand shake to miner err:->", ack.Message)
		return
	}

	pipe := NewPipe(conn, consumeConn, c.payCh)
	go pipe.collectRequest()

	pipe.pullDataFromServer()
}

func (c *Client) NewHandReq(target string) *service.HandShake {
	reqData := &service.HandShakeData{
		Addr:   c.Address.ToString(),
		Target: target,
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		fmt.Println("marshal hand shake data err:->", err)
		return nil
	}

	sig := ed25519.Sign(c.Key.PriKey, data)
	req := &service.HandShake{
		Sig:           sig,
		HandShakeData: reqData,
	}

	return req
}
