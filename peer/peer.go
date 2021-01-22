/*
Package peer provides the ability to setup connections with peers as well
as manage sending and receiving torrent pieces with those peers.
*/
package peer

import (
    "net"
    "encoding/binary"
    "strconv"
    "time"

    "github.com/pkg/errors"
)

const connTimeout = 20 * time.Second

// Errors
var (
    ErrBadPeers = errors.New("Received malformed peers list")
)

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Host net.IP
    Port uint16
    conn net.Conn
}

func (peer Peer) String() string {
    return net.JoinHostPort(peer.Host.String(), strconv.Itoa(int(peer.Port)))
}

// Unmarshal creates a list of Peers from a serialized list of peers
func Unmarshal(peersBytes []byte) ([]Peer, error) {
    if len(peersBytes) % 6 != 0 {
        return nil, errors.Wrap(ErrBadPeers, "Unmarshal")
    }

    numPeers := len(peersBytes) / 6
    peersList := make([]Peer, numPeers)

    for i := 0; i < numPeers; i++ {
        peersList[i].Host = net.IP(peersBytes[ i*6 : i*6+4 ])
        peersList[i].Port = binary.BigEndian.Uint16(peersBytes[ i*6+4 : (i+1)*6 ])
    }

    return peersList, nil
}
