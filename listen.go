package main

import (
    "bytes"

    "github.com/kylec725/graytorrent/peer"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

// Loop to listen on incoming connections for potential new peers
func peerListen() {
    for {
        conn, err := listener.Accept()
        if err != nil {  // Exit if the listener encounters an error
            if errors.Unwrap(err).Error() != "use of closed network connection" {
                log.WithField("error", err.Error()).Debug("Listener shutdown")
            }
            return
        }

        newPeer := peer.New(conn.RemoteAddr().String(), conn, nil)
        infoHash, err := newPeer.RcvHandshake()
        if err != nil {
            log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Incoming peer handshake sequence failed")
            continue
        }

        // Check if the infohash matches any torrents we are serving
        for i, to := range torrentList {
            if bytes.Equal(infoHash[:], to.Info.InfoHash[:]) {
                newPeer.Info = &torrentList[i].Info  // Assign correct info before sending handshake
                // Send back a handshake
                if err = newPeer.SendHandshake(); err != nil {
                    log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Incoming peer handshake sequence failed")
                    break
                }

                torrentList[i].IncomingPeers <- newPeer  // Send to torrent session
                log.WithField("peer", newPeer.String()).Debug("Incoming peer was accepted")
            }
        }
    }
}
