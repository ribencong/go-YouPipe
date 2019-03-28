package service

import "fmt"

const (
	SocksServerPoint = ":52018"
)

var Config = SrvConf{
	ServicePoint: SocksServerPoint,
}

type SrvConf struct {
	ServicePoint string `json:"accessPoint"`
}

func (conf SrvConf) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n",
		"ServicePoint", conf.ServicePoint)
}
