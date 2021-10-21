package torrent

import (
	"net"
	"strconv"

	"github.com/kylec725/graytorrent/internal/config"
	"github.com/kylec725/graytorrent/internal/connect"
	"github.com/kylec725/graytorrent/internal/peer"
	"github.com/kylec725/graytorrent/internal/peer/handshake"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrListener is used to close the peerListener safely
var ErrListener = errors.New("use of closed network connection")

// initListener initializes a listener and gets an open port for incoming peers
func initListener() (net.Listener, uint16, error) {
	portRange := config.GetConfig().Network.ListenerPort
	port, err := connect.OpenPort(portRange)
	if err != nil {
		log.WithFields(log.Fields{
			"portrange": portRange,
			"error":     err.Error(),
		}).Warn("No open port found in port range, using random port")
	}

	service := ":" + strconv.Itoa(int(port))
	listener, err := net.Listen("tcp", service)
	if err != nil {
		return listener, port, errors.Wrap(err, "initListener")
	}
	// Set global port
	port, err = connect.PortFromAddr(listener.Addr().String()) // Get actual port in case none in portrange were available
	if err != nil {
		return listener, port, errors.Wrap(err, "initListener")
	}
	return listener, port, nil
}

// peerListen loops to listen for incoming connections of peers
func (s *Session) peerListen() {
	// loop will exit as long as we call listener.Close()
	for {
		conn, err := s.peerListener.Accept()
		if err != nil { // Exit if the peerListener encounters an error
			if errors.Is(err, ErrListener) {
				log.WithField("error", err.Error()).Debug("Listener shutdown")
			}
			return
		}
		go s.acceptPeer(conn)
	}
}

func (s *Session) acceptPeer(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	infoHash, err := handshake.Read(conn)
	if err != nil {
		log.WithFields(log.Fields{"peer": addr, "error": err.Error()}).Debug("Error with incoming peer handshake")
		return
	}

	// Check if the infohash matches any torrents we are serving
	if to, ok := s.torrents[infoHash]; ok {
		// Check if the torrent's goroutine is running first
		if !to.Started {
			return
		}
		newPeer := peer.New(addr, conn, to.Info)
		if err := newPeer.RespondHandshake(to.Info); err != nil {
			log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Error when responding to handshake")
		}

		to.NewPeers <- newPeer // Send to torrent session
		log.WithField("peer", newPeer.String()).Debug("Incoming peer was accepted")
	}
}
