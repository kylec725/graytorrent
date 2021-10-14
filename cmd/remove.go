package cmd

import (
	"fmt"
	"os"

	"github.com/kylec725/graytorrent/internal/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&isInfoHash, "infohash", "i", false, "select a torrent with its infohash")
	removeCmd.Flags().BoolVarP(&rmFiles, "remove", "r", false, "remove the torrent's file(s)")
}

var (
	removeCmd = &cobra.Command{
		Use:   "rm",
		Short: "removes a managed torrent",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.Remove(args[0], isInfoHash, rmFiles); err != nil {
				fmt.Fprintf(os.Stderr, "Removing torrent failed: %v", err)
			}
			fmt.Println("Removed torrent")
		},
	}
)
