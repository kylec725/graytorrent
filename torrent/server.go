package torrent

import (
	"context"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/rpc"
	pb "github.com/kylec725/graytorrent/rpc"
	"github.com/pkg/errors"
)

// server.go contains implementations of the required grpc server functions

// List current managed torrents
func (s *Session) List(in *pb.Empty, stream pb.Torrent_ListServer) error {
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
func (s *Session) Remove(ctx context.Context, in *pb.SelectedTorrent) (*pb.Empty, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())
	if to, ok := s.torrentList[infoHash]; ok {
		s.RemoveTorrent(to)
	} else {
		return nil, errors.New("Torrent not found")
	}
	return &pb.Empty{}, nil
}

// Start a torrent's download/upload
func (s *Session) Start(ctx context.Context, in *pb.SelectedTorrent) (*pb.Empty, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())
	if to, ok := s.torrentList[infoHash]; ok {
		ctx := context.WithValue(ctx, common.KeyPort, s.port)
		to.Start(ctx)
	} else {
		return nil, errors.New("Torrent not found")
	}
	return &pb.Empty{}, nil
}

// Stop a torrent's download/upload
func (s *Session) Stop(ctx context.Context, in *pb.SelectedTorrent) (*pb.Empty, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())
	if to, ok := s.torrentList[infoHash]; ok {
		to.Stop()
	} else {
		return nil, errors.New("Torrent not found")
	}
	return &pb.Empty{}, nil
}
