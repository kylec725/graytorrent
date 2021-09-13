package main

import (
	pb "github.com/kylec725/graytorrent/rpc"
)

func (s *torrentServer) List(in *pb.ListRequest, list pb.Torrent_ListServer) error {
	server.Stop()
	// return pb.Torrent_ListClient{}, nil
	return nil
}
