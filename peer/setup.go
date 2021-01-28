package peer

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/connect"
	"github.com/kylec725/graytorrent/peer/handshake"
	"github.com/kylec725/graytorrent/peer/message"
	"github.com/pkg/errors"
)

// Errors
var (
	ErrBadPeers = errors.New("Received malformed peers list")
	ErrInfoHash = errors.New("Received infohash does not match")
)

// Unmarshal creates a list of Peers from a serialized list of peers
func Unmarshal(peersBytes []byte, info common.TorrentInfo) ([]Peer, error) {
	if len(peersBytes)%6 != 0 {
		return nil, errors.Wrap(ErrBadPeers, "Unmarshal")
	}

	numPeers := len(peersBytes) / 6
	peersList := make([]Peer, numPeers)

	for i := 0; i < numPeers; i++ {
		host := net.IP(peersBytes[i*6 : i*6+4])
		port := binary.BigEndian.Uint16(peersBytes[i*6+4 : i*6+6])
		addr := net.JoinHostPort(host.String(), strconv.Itoa(int(port)))
		peersList[i] = New(addr, nil, info)
	}

	return peersList, nil
}

func (p *Peer) dial() error {
	conn, err := net.Dial("tcp", p.String())
	if err != nil {
		return errors.Wrap(err, "dial")
	}
	p.Conn = &connect.Conn{Conn: conn, Timeout: peerTimeout}
	return nil
}

// Sends and receives a handshake from the peer
func (p *Peer) initHandshake(info common.TorrentInfo) error {
	h := handshake.New(info)
	if _, err := p.Conn.Write(h.Encode()); err != nil {
		return errors.Wrap(err, "initHandshake")
	}
	infoHash, err := handshake.Read(p.Conn.Conn)
	if err != nil {
		return errors.Wrap(err, "initHandshake")
	} else if !bytes.Equal(infoHash[:], info.InfoHash[:]) { // Verify the infohash
		return errors.Wrap(ErrInfoHash, "initHandshake")
	}
	// Send bitfield to the peer
	msg := message.Bitfield(info.Bitfield)
	_, err = p.Conn.Write(msg.Encode())
	return errors.Wrap(err, "initHandshake")
}
