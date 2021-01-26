/*
Package peer provides the ability to setup connections with peers as well
as manage sending and receiving torrent pieces with those peers.
Peers also handle writing pieces to file if necessary.
*/
package peer

import (
    "time"
    "math"
    "net"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/bitfield"
    "github.com/kylec725/graytorrent/peer/message"
    "github.com/kylec725/graytorrent/connect"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const peerTimeout = 120 * time.Second
const pollTimeout = 5 * time.Second
const startRate = 2  // Uses adaptive rate after first requests

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Addr string
    Conn *connect.Conn  // nil if not connected

    info *common.TorrentInfo
    bitfield bitfield.Bitfield
    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool
    rate int  // max number of outgoing requests/pieces a peer can queue
    workQueue []workPiece
    verified bool  // says whether the infohash matches or not
    shutdown bool
}

func (peer Peer) String() string {
    return peer.Addr
}

// New returns a new instantiated peer
func New(addr string, conn net.Conn, info *common.TorrentInfo) Peer {
    bitfieldSize := int(math.Ceil(float64(info.TotalPieces) / 8))
    return Peer{
        Addr: addr,
        Conn: &connect.Conn{Conn: conn, Timeout: handshakeTimeout},

        info: info,
        bitfield: make([]byte, bitfieldSize),
        amChoking: true,
        amInterested: false,
        peerChoking: true,
        peerInterested: false,
        rate: startRate,
        workQueue: []workPiece{},
        shutdown: false,
    }
}

// Shutdown lets the main goroutine signal a peer to stop working
func (peer *Peer) Shutdown() {
    peer.shutdown = true
}

// Choke notifies a peer that we are choking them
func (peer *Peer) Choke() error {  // Main should handle shutting down the peer if we have an error
    peer.amChoking = true
    msg := message.Choke()
    err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "Choke")
}

// Unchoke notifies a peer that we are not choking them
func (peer *Peer) Unchoke() error {
    peer.amChoking = false
    msg := message.Unchoke()
    err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "Unchoke")
}

// StartWork makes a peer wait for pieces to download
func (peer *Peer) StartWork(work chan int, results chan int, remove chan string, done chan bool) {
    ctxLog := log.WithField("peer", peer.String())
    peer.shutdown = false
    if !peer.verified {
        err := peer.initHandshake()
        if err != nil {
            ctxLog.WithField("error", err.Error()).Debug("Handshake failed")
            remove <- peer.String()  // Notify main to remove this peer from its list
            return
        }
    }
    ctxLog.Debug("Handshake successful")

    // Setup peer connection
    connection := make(chan []byte)
    go peer.Conn.Await(connection)
    peer.Conn.Timeout = peerTimeout

    // Work loop
    for {
        // Check if main told peer to shutdown
        if peer.shutdown {
            goto exit
        }

        // TODO what happens if no data is received, but we need to get more work? i.e. this is the only peer with
        // the needed piece, but we block because we don't receive data
        select {
        case data, ok := <-connection:
            if !ok {
                goto exit
            }
            msg := message.Decode(data)
            if err := peer.handleMessage(msg, work, results); err != nil {
                ctxLog.WithFields(log.Fields{"type": msg.String(), "size": len(msg.Payload), "error": err.Error()}).Debug("Error handling message")
                goto exit
            }
        case _, ok := <-done:
            if !ok {
                goto exit
            }
        case <-time.After(pollTimeout):  // Poll to get unstuck if no messages are received
        }

        // Find new work piece if queue is open
        if len(peer.workQueue) < peer.rate {
            select {
            case index := <-work:
                // Send the work back if the peer does not have the piece
                if !peer.bitfield.Has(index) {
                    work <- index
                    continue
                }

                // Download piece from the peer
                err := peer.downloadPiece(index)
                if err != nil {
                    ctxLog.WithFields(log.Fields{"piece index": index, "error": err.Error()}).Debug("Failed to start piece download")
                    work <- index  // Put piece back onto work channel
                    goto exit
                }
            default:  // Don't block if we can't find work
            }
        }
    }

    exit:
    for i := range peer.workQueue {
        work <- peer.workQueue[i].index
    }
    peer.Conn.Quit()  // Tell connection goroutine to exit
    remove <- peer.String()  // Notify main to remove this peer from its list
    ctxLog.Debug("Peer shutdown")
    return
}
