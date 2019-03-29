package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(mineCmd)
}

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "mining model for You Pipe Server",
	Long:  `TODO::.`,
	Run:   mineAction,
}

func mineAction(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		mineActionStatus()
		return
	}

	//switch args[0] {
	//case ''
	//}
}

func mineActionStatus() {
}
