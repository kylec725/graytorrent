package cmd

import (
	"fmt"

	"github.com/kylec725/graytorrent/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&magnetLink, "magnet", "m", "", "use a magnet link instead of a .torrent file to add a torrent")
}

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "adds a new torrent from a .torrent file or magnet link",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.Add(args[0]); err != nil {
				fmt.Println(err)
			}
		},
	}
)