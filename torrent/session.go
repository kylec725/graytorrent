package torrent

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kylec725/gray/internal/common"
	pb "github.com/kylec725/gray/rpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Session is an instance of gray
type Session struct {
	torrentList  map[[20]byte]*Torrent
	peerListener net.Listener
	port         uint16
	// server       *grpc.Server
	pb.UnimplementedTorrentServer
}

// NewSession returns a new gray session
func NewSession() (Session, error) {
	log.Info("Gray Started")
	// torrentList, err := LoadAll() // NOTE: only LoadAll if we are starting a server
	// if err != nil {
	// 	return Session{}, errors.Wrap(err, "NewSession")
	// }
	listener, port, err := initListener()
	if err != nil {
		return Session{}, errors.Wrap(err, "NewSession")
	}
	return Session{
		torrentList:  make(map[[20]byte]*Torrent),
		peerListener: listener,
		port:         port,
		// server:       grpc.NewServer(),
	}, nil
}

// Close performs clean up for a session
func (s *Session) Close() {
	for _, to := range s.torrentList {
		to.Stop()
	}
	if err := s.SaveAll(); err != nil {
		log.WithField("error", err.Error()).Debug("Problem occurred while saving torrent management data")
	}
	s.peerListener.Close()
	// s.server.Stop()
	log.Info("Gray stopped")
}

// AddTorrent adds a new torrent to be managed
func (s *Session) AddTorrent(ctx context.Context, filename string) (*Torrent, error) {
	to := Torrent{File: filename}
	if err := to.Setup(ctx); err != nil {
		return nil, errors.Wrap(err, "AddTorrent")
	}
	s.torrentList[to.InfoHash] = &to
	return &to, nil
}

// RemoveTorrent removes a currently managed torrent
func (s *Session) RemoveTorrent(to Torrent) {
	to.Stop()
	delete(s.torrentList, to.InfoHash)
	// TODO: remove save data of torrent
}

// TODO: add an option to resume a torrent if it matches this one

// Download begins a download for a single torrent
func (s *Session) Download(ctx context.Context, filename string) {
	defer s.Close()
	go s.catchSignal()
	go s.peerListen()

	ctx = context.WithValue(ctx, common.KeyPort, s.port)

	to, err := s.AddTorrent(ctx, filename)
	if err != nil {
		fmt.Println("Single torrent failed:", err)
		log.WithFields(log.Fields{"filename": filename, "error": err.Error()}).Info("Failed to add torrent")
		return
	}
	log.WithField("name", to.Info.Name).Info("Torrent added")

	go to.Start(ctx) // NOTE: maybe add an option to seed after download is complete
	for to.Info.Left > 0 {
		time.Sleep(time.Second)
	}
	fmt.Println("Torrent done:", to.Info.Name)
}

func (s *Session) catchSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	_ = <-signalChan // Cleanup on interrupt signal
	signal.Stop(signalChan)
	s.Close()
	os.Exit(0)
}
