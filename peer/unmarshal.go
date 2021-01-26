package peer

import (
    "net"
    "encoding/binary"
    "strconv"

    "github.com/kylec725/graytorrent/common"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrBadPeers = errors.New("Received malformed peers list")
)

// Unmarshal creates a list of Peers from a serialized list of peers
func Unmarshal(peersBytes []byte, info *common.TorrentInfo) ([]Peer, error) {
    if len(peersBytes) % 6 != 0 {
        return nil, errors.Wrap(ErrBadPeers, "Unmarshal")
    }

    numPeers := len(peersBytes) / 6
    peersList := make([]Peer, numPeers)

    for i := 0; i < numPeers; i++ {
        host := net.IP(peersBytes[ i*6 : i*6+4 ])
        port := binary.BigEndian.Uint16(peersBytes[ i*6+4 : (i+1)*6 ])
        addr := net.JoinHostPort(host.String(), strconv.Itoa(int(port)))
        peersList[i] = New(addr, nil, info)
    }

    return peersList, nil
}
