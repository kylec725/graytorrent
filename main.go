package main

import (
	"os"
	"path/filepath"

	"github.com/kylec725/gray/cmd"
)

var (
	grayPath = filepath.Join(os.Getenv("HOME"), ".config", "gray")
)

// Launches cobra, code execution begins in cmd/root.go
func main() {
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
