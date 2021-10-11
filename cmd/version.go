package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.20"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gray",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gray version", version)
	},
}
