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
    "github.com/kylec725/graytorrent/write"
    "github.com/kylec725/graytorrent/connect"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const pollTimeout = time.Second
const startRate = 3  // slow approach: hard limit on requests per peer
const reqSize = 16384

// Errors
var (
    ErrBadPeers = errors.New("Received malformed peers list")
)

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Host net.IP
    Port uint16
    Conn *connect.Conn  // nil if not connected
    Bitfield bitfield.Bitfield

    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool
    info *common.TorrentInfo
    rate int  // number of outgoing requests
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
        Bitfield: make([]byte, info.TotalPieces),
        info: info,
        rate: startRate,
    }
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

func (peer *Peer) adjustRate(actualRate int) {
    // Use aggressive algorithm from rtorrent
    if actualRate < 20 {
        peer.rate = actualRate + 2
    } else {
        peer.rate = actualRate / 5 + 18
    }
}

// TODO
// Requests a piece in blocks from a peer
func (peer *Peer) requestPiece(index int) ([]byte, error) {
    // start of elapsed time
    // send peer.rate number of requests for blocks
    amountLeft := common.PieceSize(peer.info, index)
    // for now, send one request at a time
    for curr := 0; amountLeft > 0; {
        peer.sendRequest(index, curr, reqSize)
        // receive request
        
        amountLeft -= reqSize
    }
    // wait for blocks to come back
    // end of elapsed time
    return nil, nil
}

// Work makes a peer wait for pieces to download
// TODO
func (peer *Peer) Work(work chan int, remove chan string) {
    err := peer.initHandshake()
    if err != nil {
        log.WithFields(log.Fields{
            "peer": peer.String(),
            "error": err.Error(),
        }).Debug("Peer handshake failed")
        remove <- peer.String()
        return
    }

    // Grab work from the channel
    for {
        // Change connection timeout to poll setting
        peer.Conn.Timeout = pollTimeout

        select {
        case index := <-work:
            fmt.Println("work index received:", index)
            if !peer.Bitfield.Has(index) {
                work <- index
                fmt.Println("piece returned:", index)
                continue
            }
            // Request piece
            piece, err := peer.requestPiece(index)
            if err != nil {
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "piece index": index,
                    "error": err.Error(),
                }).Debug("Requesting piece from peer failed")
                remove <- peer.String()
                return
            }
            // Verify the piece's hash
            if !write.VerifyPiece(peer.info, index, piece) {
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "piece index": index,
                }).Debug("Received a piece with an invalid hash")
                return  // TODO maybe retry requesting the piece until the hash matches?
            }
            // Add the piece to file
            if err = write.AddPiece(peer.info, index, piece); err != nil {
                log.WithFields(log.Fields{
                    "peer": peer.String(),
                    "piece index": index,
                    "error": err.Error(),
                }).Debug("Writing piece from peer failed")
                remove <- peer.String()
                return
            }
        default:
            fmt.Println("check for new message")
        }
    }
    // If peer disconnects, send its address to the remove channel
}
