package peer

import (
    "net"
    "time"
    "bytes"
    "fmt"

    "github.com/pkg/errors"
)

const protocol = "BitTorrent protocol"

// Errors
var (
    ErrPstr = errors.New("Got bad pstr length or value")
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
        conn, err := net.Dial("tcp", peer.String())
        if err != nil {
            return errors.Wrap(err, "sendHandshake")
        }
        conn.SetDeadline(time.Now().Add(handshakeTimeout))
    }

    // Send the handshake
    handshake := peer.newHandshake()
    bytesSent, err := peer.Conn.Write(handshake)
    if err != nil {
        return errors.Wrap(err, "sendHandshake")
    } else if bytesSent != len(handshake) {  // TODO probably will change, not sure if all bytes are guaranteed to be sent
        fmt.Println("Fix sendHandshake")
        return errors.New("Unexpected number of bytes sent")
    }

    return nil
}

func (peer *Peer) rcvHandshake() error {
    buf := make([]byte, 1)
    if _, err := peer.Conn.Read(buf); err != nil {
        return errors.Wrap(err, "RcvHandshake")
    }

    pstrLen := buf[0]
    if pstrLen == 0 {
        return errors.Wrap(ErrPstr, "RcvHandshake")
    }

    buf = make([]byte, 48 + pstrLen)
    if _, err := peer.Conn.Read(buf); err != nil {
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

// Attempts to connect to a peer if necessary
func (peer *Peer) connect() error {
    if peer.Conn == nil {
        if err := peer.sendHandshake(); err != nil {
            return errors.Wrap(err, "connect")
        } else if err = peer.rcvHandshake(); err != nil {
            return errors.Wrap(err, "connect")
        }
    }
    return nil
}

// AcceptPeer attempt to handshake with an incoming peer
func (peer *Peer) AcceptPeer() error {
    if err := peer.rcvHandshake(); err != nil {
        return errors.Wrap(err, "AcceptPeer")
    }
    if err := peer.sendHandshake(); err != nil {
        return errors.Wrap(err, "AcceptPeer")
    }
    return nil
}
