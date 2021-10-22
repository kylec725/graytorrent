package torrent

import (
	"context"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/rpc"
	pb "github.com/kylec725/graytorrent/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server.go contains implementations of the required grpc server functions

// Errors
var (
	ErrTorrentNotFound = status.Error(codes.NotFound, "Torrent not found")
)

// List current managed torrents
func (s *Session) List(ctx context.Context, in *pb.Empty) (*pb.ListReply, error) {
	torrents := make([]*pb.Torrent, 0)
	for _, to := range s.torrents {
		torrents = append(torrents,
			&pb.Torrent{
				Id:          to.ID,
				Name:        to.Info.Name,
				InfoHash:    to.Info.InfoHash[:],
				TotalLength: uint32(to.Info.TotalLength),
				Left:        uint32(to.Info.Left),
				DownRate:    uint32(to.DownRate()),
				UpRate:      uint32(to.UpRate()),
				State:       rpc.Torrent_State(to.State()),
			})
	}
	reply := pb.ListReply{Torrents: torrents}
	return &reply, nil
}

// Add a new torrent to be managed
func (s *Session) Add(ctx context.Context, in *pb.AddRequest) (*pb.Empty, error) {
	_, err := s.AddTorrent(ctx, in.GetName(), in.GetMagnet(), in.GetDirectory())
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

// Remove a torrent from being managed
func (s *Session) Remove(ctx context.Context, in *pb.RemoveRequest) (*pb.Empty, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.TorrentRequest.GetInfoHash())

	if to, ok := s.torrents[infoHash]; ok {
		s.RemoveTorrent(to, in.RmFiles)
		return &pb.Empty{}, nil
	}
	// Check ID instead
	for _, to := range s.torrents {
		if to.ID == in.TorrentRequest.GetId() {
			s.RemoveTorrent(to, in.RmFiles)
			return &pb.Empty{}, nil
		}
	}

	return nil, ErrTorrentNotFound
}

// Start a torrent's download/upload
func (s *Session) Start(ctx context.Context, in *pb.TorrentRequest) (*pb.Empty, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())

	if to, ok := s.torrents[infoHash]; ok {
		if !to.Started {
			newCtx := context.WithValue(context.Background(), common.KeyPort, s.port) // NOTE: using ctx causes to.Start() to end immediately
			go to.Start(newCtx)
		}
		return &pb.Empty{}, nil
	}
	// Check ID instead
	for _, to := range s.torrents {
		if to.ID == in.GetId() {
			if !to.Started {
				newCtx := context.WithValue(context.Background(), common.KeyPort, s.port)
				go to.Start(newCtx)
			}
			return &pb.Empty{}, nil
		}
	}

	return nil, ErrTorrentNotFound
}

// Stop a torrent's download/upload
func (s *Session) Stop(ctx context.Context, in *pb.TorrentRequest) (*pb.Empty, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())

	if to, ok := s.torrents[infoHash]; ok {
		to.Stop()
		return &pb.Empty{}, nil
	}
	// Check ID instead
	for _, to := range s.torrents {
		if to.ID == in.GetId() {
			to.Stop()
			return &pb.Empty{}, nil
		}
	}

	return nil, ErrTorrentNotFound
}

// TODO: session will send updates based on a ticker (every second)
