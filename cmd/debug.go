package cmd

import (
	"context"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/youpipe/go-node/core"
	"github.com/youpipe/go-node/pbs"
	"github.com/youpipe/go-node/thread"
	"github.com/youpipe/go-node/utils"
)

/************************************************************************
*							client part
************************************************************************/
func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(gossipDebugCmd)
	debugCmd.AddCommand(threadDebugCmd)
	debugCmd.AddCommand(sysConfDebugCmd)
	debugCmd.AddCommand(logDebugCmd)

	logDebugCmd.Flags().IntVarP(&logLevelPara, "logLevel", "l", int(logging.DEBUG),
		"log level ")
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "String debug information of node.",
	Long:  `TODO::.`,
	Run:   debug,
	Args:  cobra.MinimumNArgs(1),
}

func debug(_ *cobra.Command, _ []string) {

	msg := &pbs.EmptyRequest{}

	client := DialToCmdService()

	res, err := client.ShowNodeInfo(context.Background(), msg)
	logger.Debugf("msg:%s, err:%v", res, err)
}

var gossipDebugCmd = &cobra.Command{
	Use:   "gossip",
	Short: "gossip debug.",
	Long:  `TODO::.`,
	Run:   gossipViews,
	//Args:  cobra.MinimumNArgs(1),
}

func gossipViews(_ *cobra.Command, _ []string) {
	msg := &pbs.EmptyRequest{}

	client := DialToCmdService()

	res, _ := client.ShowGossipViews(context.Background(), msg)
	logger.Debug(res.Msg)
}

var threadDebugCmd = &cobra.Command{
	Use:   "thread",
	Short: "thread debug.",
	Long:  `TODO::.`,
	Run:   threadShow,
}

func threadShow(_ *cobra.Command, _ []string) {
	msg := &pbs.EmptyRequest{}

	client := DialToCmdService()
	res, _ := client.ShowThreadInfos(context.Background(), msg)
	logger.Debug(res.Msg)
}

var sysConfDebugCmd = &cobra.Command{
	Use:   "sysconf",
	Short: "system configuration debug.",
	Long:  `TODO::.`,
	Run:   sysConf,
	//Args:  cobra.MinimumNArgs(1),
}

func sysConf(_ *cobra.Command, _ []string) {
	msg := &pbs.EmptyRequest{}

	client := DialToCmdService()
	res, _ := client.ShowSysConf(context.Background(), msg)
	logger.Debug(res.Msg)
}

var logDebugCmd = &cobra.Command{
	Use:   "log",
	Short: "set log levels",
	Long: ` youPipe debug log [module] -l <Level>
	module: ["gossip", "core", "thread","account", "service"]  
	Level: [5=DEBUG, 4=INFO, 3=NOTICE, 2=WARNING, 1=ERROR, 0=CRITICAL`,
	Run:  logLevel,
	Args: cobra.MinimumNArgs(1),
}
var logLevelPara = 5

func logLevel(_ *cobra.Command, args []string) {
	if len(args) != 1 {
		panic("set the module of the log")
	}

	msg := &pbs.LogLevel{
		Module: args[0],
		Level:  int32(logLevelPara),
	}

	client := DialToCmdService()
	res, _ := client.SetLogLevel(context.Background(), msg)
	logger.Debug(res.Msg)
}

/************************************************************************
*							Service part
************************************************************************/

func (s *cmdService) ShowGossipViews(ctx context.Context, request *pbs.EmptyRequest) (*pbs.CommonResponse, error) {

	node := core.GetNodeInst().GossipNode

	var fmtStr string
	if node == nil {
		fmtStr = "gossip is not online"
	} else {
		fmtStr = node.String()
	}

	return &pbs.CommonResponse{Msg: fmtStr}, nil
}

func (s *cmdService) ShowThreadInfos(ctx context.Context, request *pbs.EmptyRequest) (*pbs.CommonResponse, error) {
	return &pbs.CommonResponse{Msg: thread.ShowThreadInfos()}, nil
}

func (s *cmdService) ShowSysConf(ctx context.Context, request *pbs.EmptyRequest) (*pbs.CommonResponse, error) {
	return &pbs.CommonResponse{Msg: utils.ShowConfig()}, nil
}

func (s *cmdService) SetLogLevel(ctx context.Context, request *pbs.LogLevel) (*pbs.CommonResponse, error) {
	utils.SetModuleLogLevel(logging.Level(request.Level), request.Module)
	return &pbs.CommonResponse{Msg: "success"}, nil
}
