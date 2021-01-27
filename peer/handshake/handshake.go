package handshake

import (
    "io"

    "github.com/kylec725/graytorrent/common"
    "github.com/pkg/errors"
)

const protocol = "BitTorrent protocol"

// Errors
var (
    ErrPstrLen = errors.New("Got bad pstr length")
    ErrPstr = errors.New("Got incorrect pstr")
    ErrInfoHash = errors.New("Received infohash does not match")
    // ErrPeerID = errors.New("Received peer ID was incorrect")
)

// Handshake is a message for starting the BitTorrent protocol
type Handshake struct {
    Pstr string
    InfoHash [20]byte
    PeerID [20]byte
}

// New returns a new handshake for a given torrent
func New(info common.TorrentInfo) Handshake {
    return Handshake{
        Pstr: protocol,
        InfoHash: info.InfoHash,
        PeerID: info.PeerID,
    }
}

// Encode serializes a handshake
func (h *Handshake) Encode() []byte {
    pstrLen := uint8(len(h.Pstr))
    handshake := make([]byte, 49 + pstrLen)
    handshake[0] = pstrLen
    curr := 1
    curr += copy(handshake[curr:], h.Pstr)
    curr += copy(handshake[curr:], make([]byte, 8))  // TODO Extensions
    curr += copy(handshake[curr:], h.InfoHash[:])
    curr += copy(handshake[curr:], h.PeerID[:])
    return handshake
}

// Read reads in a handshake from a stream and returns the infohash
func Read(reader io.Reader) ([20]byte, error) {
    buf := make([]byte, 1)
    if _, err := io.ReadFull(reader, buf); err != nil {
        return [20]byte{}, errors.Wrap(err, "Read")
    }

    pstrLen := buf[0]
    if pstrLen == 0 {
        return [20]byte{}, errors.Wrap(ErrPstrLen, "rcvHandshake")
    }

    buf = make([]byte, 48 + pstrLen)
    if _, err := io.ReadFull(reader, buf); err != nil {
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
