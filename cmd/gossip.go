package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(gossipCmd)
}

var gossipCmd = &cobra.Command{
	Use:   "gossip",
	Short: "String debug information of node.",
	Long:  `TODO::.`,
}
