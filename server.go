package main

import (
	"context"

	pb "github.com/kylec725/graytorrent/rpc"
)

func (s *torrentServer) List(in *pb.ListRequest, list pb.Torrent_ListServer) error {
	server.Stop()
	return nil
}

func (s *torrentServer) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddReply, error) {
	return nil, nil
}

func (s *torrentServer) Remove(ctx context.Context, in *pb.RemoveRequest) (*pb.RemoveReply, error) {
	return nil, nil
}

func (s *torrentServer) Start(ctx context.Context, in *pb.StartRequest) (*pb.StartReply, error) {
	return nil, nil
}

func (s *torrentServer) Stop(ctx context.Context, in *pb.StopRequest) (*pb.StopReply, error) {
	return nil, nil
}
