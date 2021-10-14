package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kylec725/graytorrent/torrent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolVarP(&magnet, "magnet", "m", false, "use a magnet link instead of a .torrent file to download")
	downloadCmd.Flags().StringVarP(&directory, "directory", "d", "", "specify the directory to save the torrent")
}

var (
	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "download a single torrent from a .torrent file or magnet link",
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			initLog()
		},
		Run: func(cmd *cobra.Command, args []string) {
			defer logFile.Close()
			if err := torrent.Download(context.Background(), args[0], magnet, directory); err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v", err)
			}
			fmt.Println("Download done")
		},
	}
)
