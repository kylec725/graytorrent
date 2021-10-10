package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of graytorrent",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("graytorrent v0.1")
	},
}
