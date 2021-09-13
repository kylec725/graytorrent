package main

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"github.com/kylec725/graytorrent/common"
	pb "github.com/kylec725/graytorrent/rpc"
	"github.com/kylec725/graytorrent/torrent"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	logLevel   = log.TraceLevel // InfoLevel || DebugLevel || TraceLevel
	serverPort = ":7001"
)

var (
	err     error
	logFile *os.File

	// Flags
	filename string
	verbose  bool
	port     uint16

	torrentList  []torrent.Torrent
	peerListener net.Listener

	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")
)

type torrentServer struct {
	pb.UnimplementedTorrentServer
}

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

	torrentList, err = torrent.LoadAll()
	if err != nil {
		log.WithField("error", err).Debug("Could not retrieve torrent management data")
	}
}

func main() {
	// Setup our context
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), common.KeyPort, port))

	// Cleanup
	defer func() {
		peerListener.Close()
		cancel()
		err = torrent.SaveAll(torrentList)
		if err != nil {
			log.WithField("error", err).Debug("Problem occurred while saving torrent management data")
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

	// Setup grpc server
	// TODO: Want to use TLS for encrypting communication
	serverListener, err := net.Listen("tcp", serverPort)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "port": serverPort[1:]}).Fatal("Failed to listen for rpc")
	}
	server := grpc.NewServer()
	pb.RegisterTorrentServer(server, &torrentServer{})
	if err = server.Serve(serverListener); err != nil {
		log.WithField("error", err).Debug("Issue with serving rpc client")
	}
}
