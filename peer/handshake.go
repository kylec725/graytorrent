package peer

import (
    "net"
    "time"
    "bytes"

    "github.com/kylec725/graytorrent/peer/message"
    "github.com/pkg/errors"
)

const protocol = "BitTorrent protocol"
const handshakeTimeout = 20 * time.Second

// Errors
var (
    ErrPstrLen = errors.New("Got bad pstr length")
    ErrPstr = errors.New("Got incorrect pstr")
    ErrInfoHash = errors.New("Received incorrect info hash")
    // ErrPeerID = errors.New("Received peer ID was incorrect")
)

func (peer *Peer) newHandshake() []byte {
    pstr := protocol
    pstrLen := uint8(len(pstr))
    handshake := make([]byte, 49 + pstrLen)
    handshake[0] = pstrLen
    curr := 1
    curr += copy(handshake[curr:], pstr)
    curr += copy(handshake[curr:], peer.info.InfoHash[:])
    curr += copy(handshake[curr:], peer.info.PeerID[:])
    return handshake
}

func (peer *Peer) sendHandshake() error {
    if peer.Conn == nil {
        // Start the TCP connection
        conn, err := net.DialTimeout("tcp", peer.String(), 3 * time.Second)
        if err != nil {
            return errors.Wrap(err, "sendHandshake")
        }
        peer.Conn.Conn = conn
    }

    // Send the handshake
    handshake := peer.newHandshake()
    err := peer.Conn.Write(handshake)
    return errors.Wrap(err, "sendHandshake")
}

func (peer *Peer) rcvHandshake() error {
    buf := make([]byte, 1)
    if err := peer.Conn.Read(buf); err != nil {
        return errors.Wrap(err, "RcvHandshake")
    }

    pstrLen := buf[0]
    if pstrLen == 0 {
        return errors.Wrap(ErrPstrLen, "RcvHandshake")
    }

    buf = make([]byte, 48 + pstrLen)
    if err := peer.Conn.Read(buf); err != nil {
        return errors.Wrap(err, "RcvHandshake")
    }

    pstr := string(buf[:pstrLen])
    if pstr != protocol {
        return errors.Wrap(ErrPstr, "RcvHandshake")
    }

    var infoHash [20]byte
    // var infoHash, peerID [20]byte
    copy(infoHash[:], buf[ pstrLen+8 : pstrLen+28 ])
    // copy(peerID[:], buf[ pstrLen+28 : pstrLen+48 ])  // TODO need to check for the corrent peer ID
    if !bytes.Equal(peer.info.InfoHash[:], infoHash[:]) {
        return errors.Wrap(ErrInfoHash, "RcvHandshake")
    }

    return nil
}

// Verifies a peer has sent a handshake if necessary
func (peer *Peer) verifyHandshake() error {
    if peer.Conn == nil {  // Initiate handshake
        if err := peer.sendHandshake(); err != nil {
            return errors.Wrap(err, "verifyHandshake")
        } else if err = peer.rcvHandshake(); err != nil {
            return errors.Wrap(err, "verifyHandshake")
        }
    } else {  // Receive handshake
        if err := peer.rcvHandshake(); err != nil {
            return errors.Wrap(err, "verifyHandshake")
        } else if err := peer.sendHandshake(); err != nil {
            return errors.Wrap(err, "verifyHandshake")
        }
    }
    // Send bitfield to the peer
    msg := message.Bitfield(peer.info.Bitfield)
    err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "verifyHandshake")
}
