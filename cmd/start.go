package cmd

import (
	"fmt"

	"github.com/kylec725/graytorrent/internal/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&isInfoHash, "infohash", "i", false, "select a torrent with its infohash")
}

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "starts a torrent's download/upload",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.Start(args[0], isInfoHash); err != nil {
				fmt.Println(err)
			}
		},
	}
)
