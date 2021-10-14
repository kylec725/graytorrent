package cmd

import (
	"fmt"
	"os"

	"github.com/kylec725/graytorrent/internal/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolVarP(&isInfoHash, "infohash", "i", false, "select a torrent with its infohash")
}

var (
	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stops a torrent's download/upload",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.Stop(args[0], isInfoHash); err != nil {
				fmt.Fprintln(os.Stderr, "Stopping torrent failed:", err)
			} else {
				fmt.Println("Stopped torrent")
			}
		},
	}
)
