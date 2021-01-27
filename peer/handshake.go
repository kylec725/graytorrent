package peer

import (
    "net"
    "time"
    "bytes"

    "github.com/kylec725/graytorrent/peer/message"
    "github.com/kylec725/graytorrent/connect"
    "github.com/pkg/errors"
)

const protocol = "BitTorrent protocol"
const handshakeTimeout = 20 * time.Second

// Errors
var (
    ErrPstrLen = errors.New("Got bad pstr length")
    ErrPstr = errors.New("Got incorrect pstr")
    ErrInfoHash = errors.New("Received infohash does not match")
    // ErrPeerID = errors.New("Received peer ID was incorrect")
)

func (peer *Peer) newHandshake() []byte {
    pstr := protocol
    pstrLen := uint8(len(pstr))
    handshake := make([]byte, 49 + pstrLen)
    handshake[0] = pstrLen
    curr := 1
    curr += copy(handshake[curr:], pstr)
    curr += copy(handshake[curr:], make([]byte, 8))  // TODO Extensions
    curr += copy(handshake[curr:], peer.Info.InfoHash[:])
    curr += copy(handshake[curr:], peer.Info.PeerID[:])
    return handshake
}

// SendHandshake sends a handshake message to a peer
func (peer *Peer) SendHandshake() error {
    if peer.Conn == nil {
        // Start the TCP connection
        conn, err := net.DialTimeout("tcp", peer.String(), 5 * time.Second)  // TODO fix dial issues with certain peers
        if err != nil {
            return errors.Wrap(err, "sendHandshake")
        }
        peer.Conn = &connect.Conn{Conn: conn, Timeout: handshakeTimeout}
    }

    // Send the handshake
    handshake := peer.newHandshake()
    err := peer.Conn.Write(handshake)
    return errors.Wrap(err, "sendHandshake")
}

// RcvHandshake receives a handshake from a peer and returns the infohash received
func (peer *Peer) RcvHandshake() ([20]byte, error) {
    buf := make([]byte, 1)
    if err := peer.Conn.ReadFull(buf); err != nil {
        return [20]byte{}, errors.Wrap(err, "rcvHandshake")
    }

    pstrLen := buf[0]
    if pstrLen == 0 {
        return [20]byte{}, errors.Wrap(ErrPstrLen, "rcvHandshake")
    }

    buf = make([]byte, 48 + pstrLen)
    if err := peer.Conn.ReadFull(buf); err != nil {
        return [20]byte{}, errors.Wrap(err, "rcvHandshake")
    }

    pstr := string(buf[:pstrLen])
    if pstr != protocol {
        return [20]byte{}, errors.Wrap(ErrPstr, "rcvHandshake")
    }

    var infoHash [20]byte
    // var infoHash, peerID [20]byte
    copy(infoHash[:], buf[ pstrLen+8 : pstrLen+28 ])
    // copy(peerID[:], buf[ pstrLen+28 : pstrLen+48 ])  // TODO need to check for the current peer ID

    return infoHash, nil
}

// Sends and receives a handshake from the peer
func (peer *Peer) initHandshake() error {
    if err := peer.SendHandshake(); err != nil {
        return errors.Wrap(err, "initHandshake")
    }
    infoHash, err := peer.RcvHandshake()
    if err != nil {
        return errors.Wrap(err, "initHandshake")
    } else if !bytes.Equal(peer.Info.InfoHash[:], infoHash[:]) {  // Verify the infohash
        return errors.Wrap(ErrInfoHash, "initHandshake")
    }
    // Send bitfield to the peer
    msg := message.Bitfield(peer.Info.Bitfield)
    err = peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "initHandshake")
}
