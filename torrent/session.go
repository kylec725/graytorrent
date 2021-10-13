package torrent

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/write"
	pb "github.com/kylec725/graytorrent/rpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

// Errors
var (
	ErrTorrentExists = errors.New("Torrent is already being managed")
)

// Session is an instance of gray
type Session struct {
	torrentList  map[[20]byte]*Torrent
	peerListener net.Listener
	port         uint16
	pb.UnimplementedTorrentServiceServer
}

// NewSession returns a new gray session
func NewSession() (Session, error) {
	log.Info("Graytorrent started")

	torrentList, err := LoadAll()
	if err != nil {
		return Session{}, errors.Wrap(err, "NewSession")
	}

	listener, port, err := initListener()
	if err != nil {
		return Session{}, errors.Wrap(err, "NewSession")
	}

	session := Session{
		torrentList:  torrentList,
		peerListener: listener,
		port:         port,
	}

	go session.peerListen()

	return session, nil
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

	log.Info("Graytorrent stopped")
}

// AddTorrent adds a new torrent to be managed
func (s *Session) AddTorrent(ctx context.Context, name string, magnet bool, directory string) (*Torrent, error) {
	var to Torrent
	if magnet {
		to = Torrent{Magnet: name}
	} else {
		to = Torrent{File: name}
	}
	if err := to.Init(); err != nil {
		log.WithFields(log.Fields{"name": name, "error": err.Error()}).Info("Failed to add torrent")
		return nil, err
	}

	if _, ok := s.torrentList[to.InfoHash]; ok {
		return nil, ErrTorrentExists
	}

	// Initialize files for writing
	to.Info.Directory = directory
	if directory == "" {
		to.Info.Directory = viper.GetViper().GetString("torrent.defaultpath")
	}
	if err := write.NewWrite(to.Info); err != nil { // Should fail if torrent already is being managed
		return nil, errors.Wrap(err, "AddTorrent")
	}

	s.torrentList[to.InfoHash] = &to
	log.WithFields(log.Fields{"name": to.Info.Name, "infohash": hex.EncodeToString(to.InfoHash[:])}).Info("Torrent added")
	return &to, nil
}

// RemoveTorrent removes a currently managed torrent
func (s *Session) RemoveTorrent(to *Torrent) {
	to.Stop()
	os.Remove(to.saveFile())
	delete(s.torrentList, to.InfoHash)
	log.WithFields(log.Fields{"name": to.Info.Name, "infohash": hex.EncodeToString(to.InfoHash[:])}).Info("Torrent removed")
	// TODO: remove save data of torrent
}

// Download begins a download for a single torrent
func (s *Session) Download(ctx context.Context, name string, magnet bool, directory string) {
	defer s.Close()

	go s.catchSignal()

	ctx = context.WithValue(ctx, common.KeyPort, s.port)

	to, err := s.AddTorrent(ctx, name, magnet, directory)
	if err != nil {
		fmt.Println("Single torrent failed:", err)
		log.WithFields(log.Fields{"name": name, "error": err.Error()}).Info("Failed to add torrent")
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
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	_ = <-signalChan // Cleanup on interrupt signal
	signal.Stop(signalChan)
	s.Close()
	os.Exit(0)
}
