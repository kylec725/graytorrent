package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v0.1"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of graytorrent",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("graytorrent", version)
	},
}
