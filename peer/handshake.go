package peer

import (
    "net"
    "time"
    "fmt"

    "github.com/pkg/errors"
)

const protocol = "BitTorrent protocol"

// NewHandshake creates a serialized handshake message
func (peer *Peer) NewHandshake(infoHash [20]byte, peerID [20]byte) []byte {
    pstr := protocol
    pstrLen := uint8(len(pstr))
    handshake := make([]byte, 49 + pstrLen)
    handshake[0] = pstrLen
    curr := 1
    curr += copy(handshake[curr:], pstr)
    curr += copy(handshake[curr:], infoHash[:])
    curr += copy(handshake[curr:], peerID[:])
    return handshake
}

// SendHandshake sends a handshake message to a peer
func (peer *Peer) SendHandshake(handshake []byte) error {
    // Start the TCP connection
    conn, err := net.Dial("tcp", peer.String())
    if err != nil {
        return errors.Wrap(err, "sendHandshake")
    }
    conn.SetDeadline(time.Now().Add(connTimeout))

    // Send the handshake
    bytesSent, err := conn.Write(handshake)
    if err != nil {
        return errors.Wrap(err, "sendHandshake")
    } else if bytesSent != len(handshake) {  // TODO probably will change, not sure if all bytes are guaranteed to be sent
        fmt.Println("Fix sendHandshake")
        return errors.New("Unexpected number of bytes sent")
    }

    peer.conn = conn
    return nil
}
