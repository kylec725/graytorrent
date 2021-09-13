package main

import (
	"net"

	pb "github.com/kylec725/graytorrent/grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	port = 7001
)

type torrentServer struct {
	pb.UnimplementedTorrentServer
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
}
