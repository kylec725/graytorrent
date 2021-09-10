package main

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/torrent"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
)

const logLevel = log.TraceLevel // InfoLevel || DebugLevel || TraceLevel

var (
	err     error
	logFile *os.File

	// Flags
	filename string
	verbose  bool
	port     uint16

	torrentList []torrent.Torrent
	listener    net.Listener

	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")
)

func init() {
	flag.StringVarP(&filename, "file", "f", "", "Filename of torrent file")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Print events to stdout")
	flag.Parse()

	err = os.MkdirAll(grayTorrentPath, os.ModePerm)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Could not create necessary directories")
	}

	setupLog()
	log.Info("Graytorrent started")

	setupViper()
	viper.WatchConfig()

	setupListen()
}

func main() {
	// Setup our context
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), common.KeyPort, port))

	// Cleanup
	defer func() {
		listener.Close()
		cancel()
		err = torrent.SaveAll(torrentList)
		if err != nil {
			log.WithField("error", err).Info("Problem occurred while saving torrent management data")
		}
		log.Info("Graytorrent stopped")
		logFile.Close()
	}()

	go peerListen() // Listen for incoming peer connections

	// Single file torrent then exit
	if filename != "" {
		go catchInterrupt(ctx, cancel) // Make sure cleanup still happens if interrupt signal is sent
		singleTorrent(ctx)
		return
	}

	// Initialize GUI
	// defer g.Close()

	// Send torrent stopped messages
}
