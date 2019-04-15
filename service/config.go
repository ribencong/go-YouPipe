package service

import (
	"fmt"
)

const (
	BuffSize         = 1 << 15
	DefaultKingKey   = "YP5rttHPzRsAe2RmF52sLzbBk4jpoPwJLtABaMv6qn7kVm"
	SocksServerPoint = "0.0.0.0"
)

var Config = SrvConf{
	KingKey:   DefaultKingKey,
	ServiceIP: SocksServerPoint,
}

type SrvConf struct {
	KingKey   string `json:"kingKey"`
	ServiceIP string `json:"accessPoint"`
}

func (conf SrvConf) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n",
		"ServiceIP", conf.ServiceIP)
}
