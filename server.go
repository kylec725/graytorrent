package main

import (
	"context"

	pb "github.com/kylec725/graytorrent/rpc"
)

func (s *torrentServer) List(ctx context.Context, in *pb.ListRequest) (pb.Torrent_ListClient, error) {
	server.Stop()
	// return pb.Torrent_ListClient{}, nil
	return nil, nil
}
