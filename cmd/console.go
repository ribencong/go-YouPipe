package cmd

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/core"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/service"
	"github.com/youpipe/go-youPipe/utils"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var nbsUsage = `TODO::......`

var logger, _ = logging.GetLogger("cmd")

var rootCmd = &cobra.Command{
	Use: "youPipe",

	Short: "youPipe is a new generation of vpn using block chain technology.",

	Long: nbsUsage,

	Run: mainRun,
}

var param struct {
	version    bool
	debug      bool
	confFile   string
	server     string
	bootServer string
	withMining string
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

	rootCmd.Flags().StringVarP(&param.withMining, "mine",
		"m", "", "Start server with mining function -m [PASSWORD] "+
			"will unlock your account and mine YPC")

	rootCmd.Flags().BoolVarP(&param.debug, "debug",
		"d", false, "run in debug model")
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initYouPipeConf() {

	core.LoadYouPipeConf(param.confFile)

	if len(param.server) != 0 {
		service.Config.ServicePoint = param.server
	}
	if len(param.bootServer) != 0 {
		network.Config.BootStrapServer = param.bootServer
	}
	if param.debug {
		utils.SysDebugMode = true
		utils.SystemLogLevel = logging.DEBUG
	}
	utils.ApplyLogLevel()
	logger.Info(core.ConfigShow())
}

func unlockMinerAccount() error {

	acc := account.GetAccount()
	if acc.IsEmpty() {
		return fmt.Errorf("no account, use: [youPipe account create]")
	}
	if len(param.withMining) > 0 {
		if ok := acc.UnlockAcc(param.withMining); !ok {
			return fmt.Errorf("account password wrong")
		} else {
			logger.Info("Unlock miner account success!")
			return nil
		}
	} else {
		fmt.Println("******Please unlock miner account******")
	TryAgain:
		fmt.Print("password:")
		bytePassword, _ := terminal.ReadPassword(0)

		if ok := acc.UnlockAcc(string(bytePassword)); !ok {
			fmt.Print("\n wrong! please try again\n")
			goto TryAgain
		}
	}

	return nil
}

func mainRun(_ *cobra.Command, _ []string) {

	if param.version {
		fmt.Println(utils.CurrentVersion)
		return
	}

	if err := unlockMinerAccount(); err != nil {
		fmt.Println(err.Error())
		return
	}

	go startCmdService()

	initYouPipeConf()

	node := core.GetNodeInst()
	node.Start()

	done := make(chan bool, 1)
	go waitSignal(done)
	<-done
}

func waitSignal(done chan bool) {

	pid := strconv.Itoa(os.Getpid())
	logger.Warningf("\n>>>>>>>>>>YouPipe node start at pid(%s)<<<<<<<<<<", pid)
	if err := ioutil.WriteFile(utils.SysConf.PidPath, []byte(pid), 0644); err != nil {
		fmt.Print("failed to write running pid", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh

	core.GetNodeInst().Destroy()
	logger.Warningf("\n>>>>>>>>>>process finished(%s)<<<<<<<<<<\n", sig)

	done <- true
}
