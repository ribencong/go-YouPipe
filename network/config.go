package network

import (
	"fmt"
	"net"
)

const (
	DefaultBootServer = "155.138.201.205"
)

var Config = NetConf{
	BootStrapServer: DefaultBootServer,
}

type NetConf struct {
	BootStrapServer string `json:"bootStrap"`
}

func (conf NetConf) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n",
		"BootStrapServer", conf.BootStrapServer)
}

func JoinHostPort(h string, p uint16) string {
	return net.JoinHostPort(h, fmt.Sprintf("%d", p))
}
