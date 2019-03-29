package utils

import (
	"fmt"
	"path/filepath"
)

const (
	CurrentVersion = "0.1.0"
	SysTimeFormat  = "2006-01-02 15:04:05"
	CmdServicePort = "52019"
)

func init() {
	baseDir := SysBaseDir()
	SysConf.ConfPath = filepath.Join(baseDir, string(filepath.Separator), "conf.json")
	SysConf.LogPath = filepath.Join(baseDir, string(filepath.Separator), "yp.log")
	SysConf.AccDataPath = filepath.Join(baseDir, string(filepath.Separator), "acc.data")
	SysConf.PidPath = filepath.Join(baseDir, string(filepath.Separator), "pid")
}

type SysConfig struct {
	ConfPath    string
	LogPath     string
	AccDataPath string
	PidPath     string
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
		"PidPath", c.PidPath)
}

func ShowConfig() string {
	str := fmt.Sprintf("\n\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n" +
		SysConf.String() +
		"++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
	)
	return str
}
