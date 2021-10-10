package main

import (
	"os"
	"path/filepath"

	"github.com/kylec725/graytorrent/cmd"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serverPort = ":7001" // GRPC server port
)

var (
	err error

	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")
	server          *grpc.Server
)

func main() {

	// setupListen()

	// Cleanup
	defer func() {
		log.Info("Graytorrent stopped")
		// logFile.Close()
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
