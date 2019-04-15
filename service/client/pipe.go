package client

import "net"

type Pipe struct {
	proxyConn net.Conn
	consume   *ConsumerConn
}
