package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/youpipe/go-node/account"
	"github.com/youpipe/go-node/core"
	"github.com/youpipe/go-node/gossip"
	"github.com/youpipe/go-node/network"
	"github.com/youpipe/go-node/service"
	"github.com/youpipe/go-node/utils"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var nbsUsage = `TODO::......`

var logger = utils.NewLog("cmd")

var rootCmd = &cobra.Command{
	Use: "youPipe",

	Short: "youPipe is a new generation of vpn using block chain technology.",

	Long: nbsUsage,

	Run: mainRun,
}

var param struct {
	version    bool
	confFile   string
	server     string
	bootServer string
	subDur     int
}

func init() {
	rootCmd.Flags().BoolVarP(&param.version, "version",
		"v", false, "show current version")

	rootCmd.Flags().StringVarP(&param.confFile, "configFile",
		"c", "", "shadow socks server access point")

	rootCmd.Flags().StringVarP(&param.server, "service",
		"s", "", "shadow socks server access point")

	rootCmd.Flags().StringVarP(&param.bootServer, "bootstrap",
		"b", "", "boot strap server")

	rootCmd.Flags().IntVarP(&param.subDur, "gspSubTime",
		"d", -1, "subscribe duration in seconds for gossip protocol")
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadConf() {
	acc := account.GetAccount()
	if acc.IsEmpty() {
		logger.Fatal("current node is empty and need to creat an account: youPipe account create")
	}

	core.LoadYouPipeConf(param.confFile)

	if len(param.server) != 0 {
		service.Config.ServicePoint = param.server
	}
	if len(param.bootServer) != 0 {
		network.Config.BootStrapServer = param.bootServer
	}
	if param.subDur > 0 {
		gossip.Config.GspSubDuration = time.Second * time.Duration(param.subDur)
	}

	fmt.Println(core.ConfigShow())
}

func setPid(pid string) {

	if err := ioutil.WriteFile(utils.SysConf.PidPath, []byte(pid), 0644); err != nil {
		fmt.Print("failed to write running pid", err)
	}
}

func mainRun(_ *cobra.Command, _ []string) {
	if param.version {
		fmt.Print(utils.CurrentVersion)
		return
	}

	loadConf()
	go startCmdService()

	core.GetNodeInst().Run()

	pid := strconv.Itoa(os.Getpid())
	fmt.Printf(">>>>>>>>>>node start at pid(%s)<<<<<<<<<<", pid)
	setPid(pid)

	sigCh := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		setPid("-1")
		logger.Warning(sig)
		done <- true
	}()
	utils.ApplyLogLevel()
	<-done
	core.GetNodeInst().Destroy()
	fmt.Printf(">>>>>>>>>>process finished<<<<<<<<<<")
}
