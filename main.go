package main

import (
	"os"
	"path/filepath"

	"github.com/kylec725/graytorrent/cmd"
	pb "github.com/kylec725/graytorrent/rpc"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	logLevel   = log.TraceLevel // InfoLevel || DebugLevel || TraceLevel
	serverPort = ":7001"        // GRPC server port
)

var (
	err     error
	logFile *os.File

	// Flags
	verbose bool

	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")
	server          *grpc.Server
)

type torrentServer struct {
	pb.UnimplementedTorrentServer
}

func init() {
	flag.BoolVarP(&verbose, "verbose", "v", false, "print events to stdout")
	// flag.Parse()
}

func main() {
	err = os.MkdirAll(grayTorrentPath, os.ModePerm)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Could not create necessary directories")
	}

	initLog()
	log.Info("Graytorrent started")

	initConfig()
	viper.WatchConfig()

	// setupListen()

	// Cleanup
	defer func() {
		log.Info("Graytorrent stopped")
		logFile.Close()
	}()

	cmd.Execute()

	// Setup grpc server
	// TODO: Want to use TLS for encrypting communication
	// serverListener, err := net.Listen("tcp", serverPort)
	// if err != nil {
	// 	log.WithFields(log.Fields{"error": err, "port": serverPort[1:]}).Fatal("Failed to listen for rpc")
	// }
	// server = grpc.NewServer()
	// pb.RegisterTorrentServer(server, &torrentServer{})
	// if err = server.Serve(serverListener); err != nil {
	// 	log.WithField("error", err).Debug("Issue with serving rpc client")
	// }
}
