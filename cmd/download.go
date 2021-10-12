package cmd

import (
	"context"

	"github.com/kylec725/graytorrent/torrent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&magnetLink, "magnet", "m", "", "use a magnet link instead of a .torrent file to download")
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
			session, err := torrent.NewSession()
			if err != nil {
				log.WithField("error", err).Info("Error when starting a new session for download")
			}
			session.Download(context.Background(), args[0])
		},
	}
)
