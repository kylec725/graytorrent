package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "graytorrent",
	Short: "graytorrent is a BitTorrent engine",
	Long: `An engine built to be compliant with the BitTorrent Protocol
	and allow the management of torrents.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Println("hello root command")
	},
}

// Execute runs a command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
