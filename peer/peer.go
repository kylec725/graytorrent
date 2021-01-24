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

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/bitfield"
    "github.com/kylec725/graytorrent/write"
    "github.com/kylec725/graytorrent/connect"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const pollTimeout = 2 * time.Second
const startRate uint16 = 3  // slow approach: hard limit on requests per peer

// Errors
var (
    ErrBadPeers = errors.New("Received malformed peers list")
)

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Host net.IP
    Port uint16
    Conn *connect.Conn  // nil if not connected

    info *common.TorrentInfo
    bitfield bitfield.Bitfield
    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool
    reqsOut uint16  // number of outgoing requests
    rate uint16  // max number of outgoing requests
    workLeft int  // amount of bytes left to download from a piece
    shutdown bool
}

func (peer Peer) String() string {
    return net.JoinHostPort(peer.Host.String(), strconv.Itoa(int(peer.Port)))
}

// New returns a new instantiated peer
func New(host net.IP, port uint16, conn net.Conn, info *common.TorrentInfo) Peer {
    return Peer{
        Host: host,
        Port: port,
        Conn: &connect.Conn{Conn: conn, Timeout: handshakeTimeout},  // Use a pointer so we can have a nil value

        info: info,
        bitfield: nil,
        amChoking: true,
        amInterested: false,
        peerChoking: true,
        peerInterested: false,
        reqsOut: 0,
        rate: startRate,
        workLeft: 0,
        shutdown: false,
    }
}

// Shutdown lets the main goroutine signal a peer to stop working
func (peer *Peer) Shutdown() {
    peer.shutdown = true
}

// Choke changes a peer's state of amChoking to true
// TODO
func (peer *Peer) Choke() error {
    // Send choking message
    peer.amChoking = true
    return nil
}

// Unchoke changes a peer's state of amChoking to false
// TODO
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

// StartWork makes a peer wait for pieces to download
func (peer *Peer) StartWork(work chan int, remove chan string) {
    err := peer.verifyHandshake()
    if err != nil {
        log.WithFields(log.Fields{
            "peer": peer.String(),
            "error": err.Error(),
        }).Debug("Peer handshake failed")
        remove <- peer.String()  // Notify main to remove this peer from its list
        return
    }

    // Change connection timeout to poll setting
    peer.Conn.Timeout = pollTimeout

    // Work loop
    for {
        // Check if peer should shut down
        if peer.shutdown {
            if err := peer.Conn.Close(); err != nil {
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "error": err.Error(),
                }).Debug("Error disconnecting with peer")
            }
            remove <- peer.String()  // Notify main to remove this peer from its list
            return
        }

        select {
        // Grab work from the channel
        case index := <-work:
            // Send the work back if the peer does not have the piece
            if !peer.bitfield.Has(index) {
                work <- index
                continue
            }

            // Download piece from the peer
            peer.reqsOut = 0
            piece, err := peer.downloadPiece(index)
            if err != nil {
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "piece index": index,
                    "error": err.Error(),
                }).Debug("Download piece failed")
                work <- index  // Put piece back onto work channel

                // Kill peer if issue was not the piece hash
                if errors.Unwrap(err) != ErrPieceHash {
                    peer.Shutdown()
                }
                continue
            }

            // Write piece to file
            if err = write.AddPiece(peer.info, index, piece); err != nil {
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "piece index": index,
                }).Debug("Writing piece to file failed")
                work <- index
                continue
            } else {  // Write was successful
                peer.info.Bitfield.Set(index)
                continue
            }
        default:
            // Receive a message from the peer
            msg, err := peer.getMessage()
            if _, err = peer.handleMessage(msg, nil); err != nil {  // Handle message
                if errors.Unwrap(err) == connect.ErrTimeout {
                    // Timeout on polling is fine
                    continue
                }
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "error": err.Error(),
                }).Debug("Received bad message")
                peer.Shutdown()
                continue
            }
        }
    }
}
