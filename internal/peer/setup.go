package peer

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"

	"github.com/kylec725/gray/internal/common"
	"github.com/kylec725/gray/internal/connect"
	"github.com/kylec725/gray/internal/peer/handshake"
	"github.com/kylec725/gray/internal/peer/message"
	"github.com/pkg/errors"
)

// Errors
var (
	ErrBadPeers = errors.New("Received malformed peers list")
	ErrInfoHash = errors.New("Received infohash does not match")
)

// Unmarshal creates a list of Peers from a serialized list of peers
func Unmarshal(peersBytes []byte, info *common.TorrentInfo) ([]Peer, error) {
	if len(peersBytes)%6 != 0 {
		return nil, errors.Wrap(ErrBadPeers, "Unmarshal")
	}

	numPeers := len(peersBytes) / 6
	peersList := make([]Peer, numPeers)

	for i := 0; i < numPeers; i++ {
		host := net.IP(peersBytes[i*6 : i*6+4])
		port := binary.BigEndian.Uint16(peersBytes[i*6+4 : i*6+6])
		if port == 0 { // Skip empty spaces in a UDP resposne
			return peersList[:i], nil
		}
		addr := net.JoinHostPort(host.String(), strconv.Itoa(int(port)))
		peersList[i] = New(addr, nil, info)
	}

	return peersList, nil
}

// Dial establishes a TCP connection with a peer
func (p *Peer) Dial() error {
	d := net.Dialer{Timeout: peerTimeout}
	conn, err := d.Dial("tcp", p.String())
	if err != nil {
		return errors.Wrap(err, "Dial")
	}
	p.Conn = &connect.Conn{Conn: conn, Timeout: peerTimeout}
	return nil
}

// InitHandshake sends and receives a handshake from the peer
func (p *Peer) InitHandshake(info *common.TorrentInfo) error {
	h := handshake.New(info)
	if _, err := p.Conn.Write(h.Encode()); err != nil {
		return errors.Wrap(err, "InitHandshake")
	}
	infoHash, err := handshake.Read(p.Conn.Conn)
	if err != nil {
		return errors.Wrap(err, "InitHandshake")
	} else if !bytes.Equal(infoHash[:], info.InfoHash[:]) { // Verify the infohash
		return errors.Wrap(ErrInfoHash, "InitHandshake")
	}
	// Send bitfield to the peer
	msg := message.Bitfield(info.Bitfield)
	_, err = p.Conn.Write(msg.Encode())
	return errors.Wrap(err, "InitHandshake")
}

// RespondHandshake responds to a received handshake form a peer
func (p *Peer) RespondHandshake(info *common.TorrentInfo) error {
	h := handshake.New(info)
	if _, err := p.Conn.Write(h.Encode()); err != nil {
		return errors.Wrap(err, "RespondHandshake")
	}

	// Send bitfield to the peer
	msg := message.Bitfield(info.Bitfield)
	if _, err := p.Conn.Write(msg.Encode()); err != nil {
		return errors.Wrap(err, "RespondHandshake")
	}

	p.Conn.Timeout = peerTimeout
	return nil
}
