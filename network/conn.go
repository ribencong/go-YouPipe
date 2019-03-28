package network

import "net"

type Conn struct {
	ConnId string
	net.Conn
}
