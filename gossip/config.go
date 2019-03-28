package gossip

import (
	"fmt"
	"time"
)

const (
	ServicePort      = "52020"
	DefaultGspSubDur = 24 * 60 * 60
)

var Config = GspConf{
	GspSubDuration: time.Second * DefaultGspSubDur,
}

type GspConf struct {
	GspSubDuration time.Duration `json:"expire"`
}

func (conf GspConf) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n",
		"GspSubDuration", conf.GspSubDuration)
}
