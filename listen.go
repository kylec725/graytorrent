package main

import (
    "bytes"

    "github.com/kylec725/graytorrent/peer"
    "github.com/kylec725/graytorrent/peer/handshake"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

// ErrListener is used to close the listener safely
var ErrListener = errors.New("use of closed network connection")

// Loop to listen on incoming connections for potential new peers
func peerListen() {
    for {
        conn, err := listener.Accept()
        if err != nil {  // Exit if the listener encounters an error
            if errors.Is(err, ErrListener) {
                log.WithField("error", err.Error()).Debug("Listener shutdown")
            }
            return
        }
        addr := conn.RemoteAddr().String()

        infoHash, err := handshake.Read(conn)
        if err != nil {
            log.WithFields(log.Fields{"peer": addr, "error": err.Error()}).Debug("Incoming peer handshake sequence failed")
            continue
        }

        // Check if the infohash matches any torrents we are serving
        for i, to := range torrentList {
            if bytes.Equal(infoHash[:], to.Info.InfoHash[:]) {
                newPeer := peer.New(addr, conn, to.Info)
                // Send back a handshake
                h := handshake.New(to.Info)
                if _, err = newPeer.Conn.Write(h.Encode()); err != nil {
                    log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Incoming peer handshake sequence failed")
                    break
                }

                torrentList[i].IncomingPeers <- newPeer  // Send to torrent session
                log.WithField("peer", newPeer.String()).Debug("Incoming peer was accepted")
            }
        }
    }
}
