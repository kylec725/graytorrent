package torrent

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kylec725/gray/internal/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Session is an instance of gray
type Session struct {
	torrentList  []Torrent // TODO: make torrentList a map[infohash]Torrent
	peerListener net.Listener
	port         uint16
	server       *grpc.Server
}

// NewSession returns a new gray session
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
		port:         port,
		server:       nil,
	}, nil
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

// Download begins a download for a single torrent
func (s *Session) Download(ctx context.Context, filename string) {
	go s.peerListen()
	ctx = context.WithValue(ctx, common.KeyPort, s.port)

	to, err := s.AddTorrent(ctx, filename)
	if err != nil {
		fmt.Println("Single torrent failed:", err)
		log.WithFields(log.Fields{"filename": filename, "error": err.Error()}).Info("Failed to add torrent")
		return
	}
	log.WithField("name", to.Info.Name).Info("Torrent added")

	go to.Start(ctx)
	for to.Info.Left > 0 {
		time.Sleep(time.Second)
	}
	to.Stop()
	SaveAll(s.torrentList)
	fmt.Println("Torrent done:", to.Info.Name)
}

// func catchInterrupt(ctx context.Context, cancel context.CancelFunc) {
// 	signalChan := make(chan os.Signal, 1)
// 	signal.Notify(signalChan, os.Interrupt)
// 	select {
// 	case <-signalChan: // Cleanup on interrupt signal
// 		signal.Stop(signalChan)
// 		peerListener.Close()
// 		cancel()
// 		err = torrent.SaveAll(torrentList)
// 		if err != nil {
// 			log.WithField("error", err).Debug("Problem occurred while saving torrent management data")
// 		}
// 		log.Info("Gray stopped")
// 		logFile.Close()
// 		os.Exit(1)
// 	case <-ctx.Done():
// 	}
// }
