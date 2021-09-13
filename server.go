package main

import (
	"context"

	pb "github.com/kylec725/graytorrent/rpc"
)

func (s *torrentServer) Quit(ctx context.Context, in *pb.QuitRequest) (*pb.QuitReply, error) {
	server.Stop()
	return &pb.QuitReply{Reply: true}, nil
}
