package cmd

import (
	"github.com/kylec725/graytorrent/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var (
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "launches a terminal interface client",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Run()
		},
	}
)
