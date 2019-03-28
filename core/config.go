package core

import (
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/gossip"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/service"
	"github.com/youpipe/go-youPipe/utils"
	"io/ioutil"
	"os"
)

type YouPipeConf struct {
	CurrentVer      string `json:"currentVer"`
	LogLevel        string `json:"logLevel"`
	service.SrvConf `json:"service"`
	gossip.GspConf  `json:"gossip"`
	network.NetConf `json:"network"`
}

var defaultYouPipeConf = YouPipeConf{
	CurrentVer: utils.CurrentVersion,
	LogLevel:   utils.DefaultSystemLogLevel.String(),
	SrvConf:    service.Config,
	GspConf:    gossip.Config,
	NetConf:    network.Config,
}

func InitYouPipeConf() {

	path := utils.SysConf.ConfPath
	if _, ok := utils.FileExists(path); ok {
		return
	}

	defaultConfData, err := json.Marshal(defaultYouPipeConf)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(path, defaultConfData, 0644)
	if err != nil {
		panic(err)
	}
}

func LoadYouPipeConf(path string) {
	if len(path) == 0 {
		path = utils.SysConf.ConfPath
	}

	fil, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fil.Close()
	conf := &defaultYouPipeConf
	parser := json.NewDecoder(fil)
	if err = parser.Decode(conf); err != nil {
		panic(err)
	}
	utils.SystemLogLevel, _ = logging.LogLevel(conf.LogLevel)
	service.Config = conf.SrvConf
	gossip.Config = conf.GspConf
	network.Config = conf.NetConf
}

func ConfigShow() string {
	return fmt.Sprintf("\n\n+++++++++++++++++++++++%8s++++++++++++++++++++++++++++\n"+
		service.Config.String()+
		gossip.Config.String()+
		network.Config.String()+
		"+NodeID:%49s+\n"+
		"+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++",
		defaultYouPipeConf.CurrentVer,
		GetNodeInst().NodeId)
}

//TODO::
func MigrateConf() {
}
