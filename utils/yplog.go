package utils

import (
	"fmt"
	"github.com/op/go-logging"
	"os"
)

const (
	LMGossip              = "gossip"
	LMCore                = "core"
	LMThread              = "thread"
	LMAccount             = "account"
	LMService             = "service"
	DefaultSystemLogLevel = logging.INFO
)

var (
	SystemLogLevel = DefaultSystemLogLevel

	LogModules = []string{
		LMGossip,
		LMCore,
		LMThread,
		LMAccount,
		LMService,
	}
)

func NewLog(module string) *logging.Logger {

	logIns, _ := logging.GetLogger(module)

	logFile, err := os.OpenFile(SysConf.LogPath,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		panic(err)
	}

	fileBackend := logging.NewLogBackend(logFile, "-->", 0)

	fileFormat := logging.MustStringFormatter(
		`{time:01-02/15:04:05} %{longfunc:-30s} %{shortfile:-22.20s} > %{level:.4s} %{message}`,
	)
	fileFormatBackend := logging.NewBackendFormatter(fileBackend, fileFormat)

	leveledFileBackend := logging.AddModuleLevel(fileFormatBackend)

	cmdFormat := logging.MustStringFormatter(
		`%{color}%{time:01-02/15:04:05} %{shortfile:-20.18s} %{shortfunc:-20.20s} [%{level:.4s}] %{message}%{color:reset}`,
	)
	cmdBackend := logging.NewLogBackend(os.Stdout, "\n>>>", 0)
	formattedCmdBackend := logging.NewBackendFormatter(cmdBackend, cmdFormat)

	logging.SetBackend(leveledFileBackend, formattedCmdBackend)

	SetModuleLogLevel(SystemLogLevel, module)

	return logIns
}

type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}

func ApplyLogLevel() {
	for _, m := range LogModules {
		SetModuleLogLevel(SystemLogLevel, m)
	}

	fmt.Println()
	fmt.Println(">>>>>>>>>>Apply log level success!<<<<<<<<<<")
}

func SetModuleLogLevel(level logging.Level, module string) {
	logging.SetLevel(level, module)
}
