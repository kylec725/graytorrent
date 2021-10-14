package cmd

import (
	"fmt"
	"os"

	"github.com/kylec725/graytorrent/internal/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVarP(&magnet, "magnet", "m", false, "use a magnet link instead of a .torrent file to add a torrent")
	addCmd.Flags().StringVarP(&directory, "directory", "d", "", "specify the directory to save the torrent")
}

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "adds a new torrent from a .torrent file or magnet link",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.Add(args[0], magnet, directory); err != nil {
				fmt.Fprintf(os.Stderr, "Adding torrent failed: %v", err)
			}
			fmt.Println("Added torrent")
		},
	}
)
