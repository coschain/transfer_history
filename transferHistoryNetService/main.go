package main

import (
	"github.com/coschain/cobra"
	"os"
	"transfer_history/transferHistoryNetService/commands"
)

var rootCmd = &cobra.Command{
	Use:   "transferNet",
	Short: "transfer history net is a http service to get transfer history of cos account",
}

func addCommands() {
	rootCmd.AddCommand(commands.StartCmd())
}

func main()  {
	addCommands()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
