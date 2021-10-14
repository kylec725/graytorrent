package cmd

import (
	"fmt"

	"github.com/kylec725/graytorrent/internal/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(monitorCmd)
}

var (
	monitorCmd = &cobra.Command{
		Use:   "mon",
		Short: "monitor the managed torrents",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.Monitor(); err != nil {
				fmt.Println(err)
			}
		},
	}
)
