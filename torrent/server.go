package torrent

import (
	"context"
	"strconv"

	"github.com/kylec725/graytorrent/rpc"
	pb "github.com/kylec725/graytorrent/rpc"
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
		stream.Send(&pb.TorrentInfo{
			Name:        to.Info.Name,
			InfoHash:    to.InfoHash[:], // NOTE: may need to check if infohash is set first
			TotalLength: uint32(to.Info.TotalLength),
			Left:        uint32(to.Info.Left),
			DownRate:    uint32(to.DownRate()),
			UpRate:      uint32(to.UpRate()),
			State:       rpc.TorrentInfo_State(to.State()),
		})
	}
	return nil
}

// Add a new torrent to be managed
func (s *Session) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddReply, error) {
	to, err := s.AddTorrent(ctx, in.File)
	if err != nil {
		return nil, err
	}
	return &pb.AddReply{
		Name:     to.Info.Name,
		InfoHash: to.InfoHash[:],
	}, err
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
