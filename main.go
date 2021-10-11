package main

import (
	"github.com/kylec725/gray/cmd"
)

// Launches cobra, code execution begins in cmd/root.go
func main() {
	cmd.Execute()

	// Setup grpc server
	// TODO: Want to use TLS for encrypting communication
	// serverListener, err := net.Listen("tcp", serverPort)
	// if err != nil {
	// 	log.WithFields(log.Fields{"error": err.Error(), "port": serverPort[1:]}).Fatal("Failed to listen for rpc")
	// }
	// server = grpc.NewServer()
	// pb.RegisterTorrentServer(server, &torrentServer{})
	// if err = server.Serve(serverListener); err != nil {
	// 	log.WithField("error", err).Debug("Issue with serving rpc client")
	// }
}
