package service

import "fmt"

const (
	buffSize         = 1 << 15
	SocksServerPoint = "0.0.0.0"
)

var Config = SrvConf{
	ServiceIP: SocksServerPoint,
}

type SrvConf struct {
	ServiceIP string `json:"accessPoint"`
}

func (conf SrvConf) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n",
		"ServiceIP", conf.ServiceIP)
}
