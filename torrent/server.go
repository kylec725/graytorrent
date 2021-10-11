package torrent

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/kylec725/gray/rpc"
	viper "github.com/spf13/viper"
)

var (
	serverAddr = ":" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
)

// type torrentServer struct {
// 	pb.UnimplementedTorrentServer
// }

// List current managed torrents
func (s *Session) List(in *pb.ListRequest, stream pb.Torrent_ListServer) error {
	for _, to := range s.torrentList {
		stream.Send(&pb.TorrentInfo{})
		fmt.Println(to.Info.Name)
	}
	return nil
}

// Add a new torrent to be managed
func (s *Session) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddReply, error) {
	time := pb.AddReply{
		Message: time.Now().Format("01-02-2006 15:04:05"),
	}
	return &time, nil
}

// Remove a torrent from being managed
func (s *Session) Remove(ctx context.Context, in *pb.RemoveRequest) (*pb.RemoveReply, error) {
	return nil, nil
}

// Start a torrent's download/upload
func (s *Session) Start(ctx context.Context, in *pb.StartRequest) (*pb.StartReply, error) {
	return nil, nil
}

// Stop a torrent's download/upload
func (s *Session) Stop(ctx context.Context, in *pb.StopRequest) (*pb.StopReply, error) {
	return nil, nil
}
