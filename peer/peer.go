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
    "fmt"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/bitfield"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const connTimeout = 20 * time.Second

// Errors
var (
    ErrBadPeers = errors.New("Received malformed peers list")
    // ErrSend = errors.New("Unexpected number of bytes sent")
    // ErrRcv = errors.New("Unexpected number of bytes received")
)

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Host net.IP
    Port uint16
    Conn net.Conn  // nil if not connected
    Bitfield bitfield.Bitfield
    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool

    info *common.TorrentInfo
}

func (peer Peer) String() string {
    return net.JoinHostPort(peer.Host.String(), strconv.Itoa(int(peer.Port)))
}

// New returns a new instantiated peer
func New(host net.IP, port uint16, conn net.Conn, info *common.TorrentInfo) Peer {
    return Peer{
        Host: host,
        Port: port,
        Conn: conn,
        info: info,
        Bitfield: make([]byte, info.TotalPieces),
    }
}

// Choke changes a peer's state of amChoking to true
func (peer *Peer) Choke() error {
    // Send choking message
    peer.amChoking = true
    return nil
}

// Unchoke changes a peer's state of amChoking to false
func (peer *Peer) Unchoke() error {
    // Send unchoking message
    peer.amChoking = false
    return nil
}

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
        peersList[i] = New(host, port, nil, info)
    }

    return peersList, nil
}

// Work makes a peer wait for pieces to download
func (peer *Peer) Work(work chan int, quit chan string) {
    // Connect peer if necessary
    if peer.Conn == nil {
        if err := peer.sendHandshake(); err != nil {
            log.WithFields(log.Fields{
                "peer": peer.String(),
                "error": err.Error(),
            }).Debug("Peer handshake failed")
        } else if err = peer.rcvHandshake(); err != nil {
            log.WithFields(log.Fields{
                "peer": peer.String(),
                "error": err.Error(),
            }).Debug("Peer handshake failed")
        }
    }

    // Grab work from the channel
    for {
        select {
        case index := <-work:
            fmt.Println("work index received:", index)
            if !peer.Bitfield.Has(index) {
                work <- index
                fmt.Println("piece returned:", index)
                continue
            }
            // Request piece
        default:
            fmt.Println("check for new message")
        }
    }
    // If peer disconnects, send its address to the quit channel
}
