package client

import (
	"fmt"
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

}
