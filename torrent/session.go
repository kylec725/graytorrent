package torrent

import (
	"context"
	"encoding/hex"
	"net"
	"os"
	"os/signal"
	"path/filepath"
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

	s := Session{
		torrentList:  torrentList,
		peerListener: listener,
		port:         port,
	}

	go s.peerListen()

	return s, nil
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

	if _, ok := s.torrentList[to.Info.InfoHash]; ok {
		log.WithFields(log.Fields{"name": name, "error": ErrTorrentExists.Error()}).Info("Failed to add torrent")
		return nil, errors.Wrap(ErrTorrentExists, "AddTorrent")
	}

	// Initialize files for writing
	to.Info.Directory = directory
	if directory == "" { // If the client does not specify a directory, we use the default path
		to.Info.Directory = viper.GetViper().GetString("torrent.defaultpath")
	}
	absDir, err := filepath.Abs(to.Info.Directory)
	if err != nil {
		return nil, errors.Wrap(err, "AddTorrent")
	}
	to.Info.Directory = absDir
	if err := write.NewWrite(to.Info); err != nil { // Should fail if torrent already is being managed
		return nil, errors.Wrap(err, "AddTorrent")
	}

	s.torrentList[to.Info.InfoHash] = &to
	log.WithFields(log.Fields{"name": to.Info.Name, "infohash": hex.EncodeToString(to.Info.InfoHash[:])}).Info("Torrent added")
	return &to, nil
}

// RemoveTorrent removes a currently managed torrent
func (s *Session) RemoveTorrent(to *Torrent, rmFiles bool) {
	to.Stop()
	os.Remove(to.saveFile())
	if rmFiles {
		fullPath := filepath.Join(to.Info.Directory, to.Info.Name)
		if err := os.RemoveAll(fullPath); err != nil {
			log.WithFields(log.Fields{"name": to.Info.Name, "infohash": hex.EncodeToString(to.Info.InfoHash[:])}).Info("Error when removing torrent's file(s)")
		}
	}
	delete(s.torrentList, to.Info.InfoHash)
	log.WithFields(log.Fields{"name": to.Info.Name, "infohash": hex.EncodeToString(to.Info.InfoHash[:])}).Info("Torrent removed")
}

// Download begins a download for a single torrent
func Download(ctx context.Context, name string, magnet bool, directory string) error {
	log.Info("Graytorrent started")
	listener, port, err := initListener()
	if err != nil {
		return err
	}

	s := Session{
		torrentList:  make(map[[20]byte]*Torrent),
		peerListener: listener,
		port:         port,
	}

	go s.peerListen()
	defer s.peerListener.Close()
	go s.catchSignal()
	ctx = context.WithValue(ctx, common.KeyPort, s.port)

	to, err := s.AddTorrent(ctx, name, magnet, directory)
	if err != nil {
		log.WithFields(log.Fields{"name": name, "error": err.Error()}).Info("Failed to add torrent")
		return err
	}

	go to.Start(ctx) // NOTE: maybe add an option to seed after download is complete
	for to.Info.Left > 0 {
		time.Sleep(time.Second)
	}

	if err := to.Save(); err != nil {
		return err
	}

	log.Info("Graytorrent stopped")
	return nil
}

func (s *Session) catchSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	_ = <-signalChan // Cleanup on interrupt signal
	signal.Stop(signalChan)
	s.Close()
	os.Exit(0)
}
