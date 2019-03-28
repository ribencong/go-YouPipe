package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/core"
	"github.com/youpipe/go-youPipe/utils"
	"golang.org/x/crypto/ssh/terminal"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Account associated operations.",
	Long:  `TODO::.`,
	Run:   showAccount,
}

var accountCreate = &cobra.Command{
	Use:   "create",
	Short: "Create an account ",
	Long:  `TODO::.`,
	Run:   createAccount,
	//Args:  cobra.MinimumNArgs(1),
}

var password string

func init() {

	rootCmd.AddCommand(accountCmd)

	accountCmd.AddCommand(accountCreate)

	accountCreate.Flags().StringVarP(&password, "password", "p", "", "pass word to create an account")
}

func createAccount(_ *cobra.Command, _ []string) {

	if !account.GetAccount().IsEmpty() {
		logger.Fatalf("Duplicate account!")
		return
	}
	if len(password) == 0 {
		var pw2 string

	TryAgain:
		fmt.Print("\nInput the pw:")
		bytePassword, _ := terminal.ReadPassword(0)
		password = string(bytePassword)

		//TODO::use password model
		if !utils.CheckAccountPassword(password) {
			goto TryAgain
		}

		fmt.Println("\nInput again:")
		bytePassword, _ = terminal.ReadPassword(0)
		pw2 = string(bytePassword)

		if password != pw2 {
			fmt.Print("Pass words are not same for the 2 times input")
			goto TryAgain
		}
	} else {
		if !utils.CheckAccountPassword(password) {
			return
		}
	}

	core.InitYouPipeConf()

	nodeId := account.CreateAccount(password)

	fmt.Printf("\nAccount(%s) create success!\n\n", nodeId)
}

func showAccount(_ *cobra.Command, _ []string) {

	if account.GetAccount().IsEmpty() {
		logger.Warning("No account to show")
		return
	}

	strShow := account.GetAccount().FormatShow()
	logger.Info(strShow)
}
