/*
Package peer provides the ability to setup connections with peers and
manage sending and receiving torrent pieces with those peers
*/
package peer

import (
    "net"
    "encoding/binary"
    "strconv"

    "github.com/pkg/errors"
)

// Errors
var (
    ErrBadPeers = errors.New("Received malformed peers list")
)

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Host net.IP
    Port uint16
}

func (p Peer) String() string {
    return net.JoinHostPort(p.Host.String(), strconv.Itoa(int(p.Port)))
}

// Unmarshal creates a list of Peers from a serialized list of peers
func Unmarshal(peersBytes []byte) ([]Peer, error) {
    if len(peersBytes) % 6 != 0 {
        return nil, errors.Wrap(ErrBadPeers, "Unmarshal")
    }

    numPeers := len(peersBytes) / 6
    peerList := make([]Peer, numPeers)

    for i := 0; i < numPeers; i++ {
        peerList[i].Host = net.IP(peersBytes[ i*6 : i*6+4 ])
        peerList[i].Port = binary.BigEndian.Uint16(peersBytes[ i*6+4 : (i+1)*6 ])
    }

    return peerList, nil
}
