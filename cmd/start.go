package cmd

import (
	"fmt"

	"github.com/kylec725/graytorrent/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "starts a torrent's download/upload",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.Start(args[0]); err != nil {
				fmt.Println(err)
			}
		},
	}
)
