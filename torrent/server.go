package torrent

import (
	"context"

	pb "github.com/kylec725/graytorrent/rpc"
)

type torrentServer struct {
	pb.UnimplementedTorrentServer
}

func (s *torrentServer) List(in *pb.ListRequest, stream pb.Torrent_ListServer) error {
	// for i := range torrentList {
	// 	stream.Send(&pb.TorrentInfo{})
	// 	fmt.Println(torrentList[i].Info.Name)
	// }
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
