package torrent

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	err     error
	logFile *os.File

	// Flags
	verbose bool

	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")
)

// Session is an instance of graytorrent
type Session struct {
	torrentList  []Torrent // TODO: make torrentList a map[infohash]Torrent
	peerListener net.Listener
	server       *grpc.Server
	port         uint16
}

// Close performs clean up for a session
func (s *Session) Close() {
	err := SaveAll(s.torrentList)
	if err != nil {
		panic("SaveAll failed")
	}
	s.peerListener.Close()
	s.server.Stop()
}

// NewSession returns a new graytorrent session
func NewSession() (Session, error) {
	torrentList, err := LoadAll()
	if err != nil {
		return Session{}, errors.Wrap(err, "NewSession")
	}
	listener, port, err := initListener()
	if err != nil {
		return Session{}, errors.Wrap(err, "NewSession")
	}
	return Session{
		torrentList:  torrentList,
		peerListener: listener,
		server:       nil,
		port:         port,
	}, nil
}

// AddTorrent adds a new torrent to be managed
func (s *Session) AddTorrent(ctx context.Context, filename string) (*Torrent, error) {
	to := Torrent{Path: filename}
	if err := to.Setup(ctx); err != nil {
		return nil, errors.Wrap(err, "AddTorrent")
	}
	s.torrentList = append(s.torrentList, to)
	return &s.torrentList[len(s.torrentList)-1], nil
}

// RemoveTorrent removes a currently managed torrent
func (s *Session) RemoveTorrent(to Torrent) {
	to.cancel()
	for i := range s.torrentList {
		if to.Path == s.torrentList[i].Path {
			s.torrentList[i] = s.torrentList[len(s.torrentList)-1]
			s.torrentList = s.torrentList[:len(s.torrentList)-1]
			return
		}
	}
}

// StartTorrent starts the download or upload of a torrent
// func StartTorrent(to Torrent) {
// 	go to.Start()
// }

// StopTorrent stops the download or upload of a torrent
func StopTorrent(to Torrent) {
	to.cancel()
}

// Download begins a download for a single torrent
func (s *Session) Download(ctx context.Context, filename string) {
	ctx = context.WithValue(ctx, common.KeyPort, s.port)
	to, err := s.AddTorrent(ctx, filename)
	if err != nil {
		fmt.Println("Single torrent failed:", err)
		log.WithFields(log.Fields{"filename": filename, "error": err.Error()}).Info("Failed to add torrent")
		return
	}
	log.WithField("name", to.Info.Name).Info("Torrent added")
	go to.Start(ctx)
	for to.Info.Left > 0 { // Go compiler marks this as data race, not a big deal, we're just polling the value
		time.Sleep(time.Second)
	}
	SaveAll(s.torrentList)
	fmt.Println("Torrent done:", to.Info.Name)
}
