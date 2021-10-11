package torrent

import (
	"bytes"
	"net"
	"strconv"

	"github.com/kylec725/graytorrent/internal/connect"
	"github.com/kylec725/graytorrent/internal/peer"
	"github.com/kylec725/graytorrent/internal/peer/handshake"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

// ErrListener is used to close the peerListener safely
var ErrListener = errors.New("use of closed network connection")

// initListener initializes a listener and gets an open port for incoming peers
func initListener() (net.Listener, uint16, error) {
	portRange := viper.GetViper().GetIntSlice("network.portrange")
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
		addr := conn.RemoteAddr().String()

		infoHash, err := handshake.Read(conn)
		if err != nil {
			log.WithFields(log.Fields{"peer": addr, "error": err.Error()}).Debug("Error with incoming peer handshake")
			continue
		}

		// Check if the infohash matches any torrents we are serving
		for i := range s.torrentList {
			// Check if the torrent's goroutine is running first
			if !s.torrentList[i].Started {
				continue
			}
			if bytes.Equal(infoHash[:], s.torrentList[i].Info.InfoHash[:]) {
				newPeer := peer.New(addr, conn, s.torrentList[i].Info)
				if err := newPeer.RespondHandshake(s.torrentList[i].Info); err != nil {
					log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Error when responding to handshake")
				}

				s.torrentList[i].NewPeers <- newPeer // Send to torrent session
				log.WithField("peer", newPeer.String()).Debug("Incoming peer was accepted")
			}
		}
	}
}
