package cmd

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/utils"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Key associated operations.",
	Long:  `TODO::.`,
}

var keyCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a key pair ",
	Long:  `TODO::.`,
	Run:   createKey,
	//Args:  cobra.MinimumNArgs(1),
}

func init() {

	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(keyCreate)

	keyCreate.Flags().StringVarP(&password, "password", "p", "", "pass word to encrypt private key")
	keyCreate.Flags().BoolVarP(&bareKey, "rawKey", "r", false, "show raw key pair")
}

var bareKey bool

func createKey(_ *cobra.Command, _ []string) {
	if !utils.CheckAccountPassword(password) {
		return
	}

	k, e := account.GenerateKey(password)
	if e != nil {
		fmt.Print(e)
	}

	if bareKey {
		fmt.Printf(" private key:[%0x]\n public key:[%0x]\n", k.PriKey, k.PubKey)
	} else {
		fmt.Printf(" cipher:[%s]\n address:[%s]\n", base58.Encode(k.LockedKey), k.ToNodeId())
	}
}
