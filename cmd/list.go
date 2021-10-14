package cmd

import (
	"fmt"
	"os"

	"github.com/kylec725/graytorrent/internal/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var (
	listCmd = &cobra.Command{
		Use:   "ls",
		Short: "list the currently managed torrents",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.List(); err != nil {
				fmt.Fprintln(os.Stderr, "Listing torrents failed:", err)
			}
		},
	}
)
