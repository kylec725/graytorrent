package cmd

import (
	"context"

	"github.com/kylec725/graytorrent/torrent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCommand)
}

var downloadCommand = &cobra.Command{
	Use:   "download",
	Short: "download a single torrent from a .torrent file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		session, err := torrent.NewSession()
		if err != nil {
			log.WithField("error", err).Info("Error when starting a new session for download")
		}
		session.Download(context.Background(), args[0])
	},
}