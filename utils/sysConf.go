package utils

import (
	"fmt"
	"path/filepath"
	"time"
)

const (
	CurrentVersion = "0.1.0"
	SysTimeFormat  = "2006-01-02 15:04:05"
	CmdServicePort = "52019"
)

var SystemTimeLoc, _ = time.LoadLocation("Asia/Shanghai")

func InitUtils() {
	baseDir := SysBaseDir()
	SystemTimeLoc, _ = time.LoadLocation("Asia/Shanghai")

	SysConf.ConfPath = filepath.Join(baseDir, string(filepath.Separator), "conf.json")
	SysConf.LogPath = filepath.Join(baseDir, string(filepath.Separator), "yp.log")
	SysConf.AccDataPath = filepath.Join(baseDir, string(filepath.Separator), "acc.data")
	SysConf.PidPath = filepath.Join(baseDir, string(filepath.Separator), "pid")
	SysConf.ReceiptPath = filepath.Join(baseDir, string(filepath.Separator), "receipt")

	initLog()
}

type SysConfig struct {
	ConfPath    string
	LogPath     string
	AccDataPath string
	PidPath     string
	ReceiptPath string
}

var SysConf = &SysConfig{}

func (c SysConfig) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n"+
		"+%-15s:%40s+\n"+
		"+%-15s:%40s+\n"+
		"+%-15s:%40s+\n",
		"LogPath", c.LogPath,
		"AccDataPath", c.AccDataPath,
		"PidPath", c.PidPath,
		"ReceiptPath", c.ReceiptPath)
}

func ShowConfig() string {
	str := fmt.Sprintf("\n\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n" +
		SysConf.String() +
		"++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
	)
	return str
}
