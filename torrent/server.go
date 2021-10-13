package torrent

import (
	"context"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/rpc"
	pb "github.com/kylec725/graytorrent/rpc"
	"github.com/pkg/errors"
)

// server.go contains implementations of the required grpc server functions

// Errors
var (
	ErrTorrentStarted  = errors.New("Torrent already started")
	ErrTorrentNotFound = errors.New("Torrent not found")
)

// List current managed torrents
func (s *Session) List(ctx context.Context, in *pb.Empty) (*pb.ListReply, error) {
	torrentList := make([]*pb.Torrent, 0)
	for _, to := range s.torrentList {
		torrentList = append(torrentList,
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
	reply := pb.ListReply{TorrentList: torrentList}
	return &reply, nil
}

// Add a new torrent to be managed
func (s *Session) Add(ctx context.Context, in *pb.AddRequest) (*pb.TorrentReply, error) {
	to, err := s.AddTorrent(ctx, in.GetName(), in.GetMagnet(), in.GetDirectory())
	if err != nil {
		return nil, err
	}
	return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
}

// Remove a torrent from being managed
func (s *Session) Remove(ctx context.Context, in *pb.RemoveRequest) (*pb.TorrentReply, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.TorrentRequest.GetInfoHash())

	if to, ok := s.torrentList[infoHash]; ok {
		s.RemoveTorrent(to, in.RmFiles)
		return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
	}
	// Check ID instead
	for _, to := range s.torrentList {
		if to.ID == in.TorrentRequest.GetId() {
			s.RemoveTorrent(to, in.RmFiles)
			return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
		}
	}

	return nil, ErrTorrentNotFound
}

// Start a torrent's download/upload
func (s *Session) Start(ctx context.Context, in *pb.TorrentRequest) (*pb.TorrentReply, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())

	if to, ok := s.torrentList[infoHash]; ok {
		if !to.Started {
			newCtx := context.WithValue(context.Background(), common.KeyPort, s.port) // NOTE: using ctx causes to.Start() to end immediately
			go to.Start(newCtx)
		}
		return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
	}
	// Check ID instead
	for _, to := range s.torrentList {
		if to.ID == in.GetId() {
			if !to.Started {
				newCtx := context.WithValue(context.Background(), common.KeyPort, s.port)
				go to.Start(newCtx)
			}
			return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
		}
	}

	return nil, ErrTorrentNotFound
}

// Stop a torrent's download/upload
func (s *Session) Stop(ctx context.Context, in *pb.TorrentRequest) (*pb.TorrentReply, error) {
	var infoHash [20]byte
	copy(infoHash[:], in.GetInfoHash())

	if to, ok := s.torrentList[infoHash]; ok {
		to.Stop()
		return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
	}
	// Check ID instead
	for _, to := range s.torrentList {
		if to.ID == in.GetId() {
			to.Stop()
			return &pb.TorrentReply{Id: to.ID, Name: to.Info.Name, InfoHash: to.Info.InfoHash[:]}, nil
		}
	}

	return nil, ErrTorrentNotFound
}
