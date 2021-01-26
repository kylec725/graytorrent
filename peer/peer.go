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
    Info *common.TorrentInfo

    bitfield bitfield.Bitfield
    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool
    rate int  // max number of outgoing requests/pieces a peer can queue
    workQueue []workPiece
}

func (peer Peer) String() string {
    return peer.Addr
}

// New returns a new instantiated peer
func New(addr string, conn net.Conn, info *common.TorrentInfo) Peer {
    var peerConn *connect.Conn = nil
    if conn != nil {
        peerConn = &connect.Conn{Conn: conn, Timeout: handshakeTimeout}
    }
    bitfieldSize := int(math.Ceil(float64(info.TotalPieces) / 8))
    return Peer{
        Addr: addr,
        Conn: peerConn,
        Info: info,

        bitfield: make([]byte, bitfieldSize),
        amChoking: true,
        amInterested: false,
        peerChoking: true,
        peerInterested: false,
        rate: startRate,
        workQueue: []workPiece{},
    }
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
    if peer.Conn == nil {
        err := peer.initHandshake()
        if err != nil {
            ctxLog.WithField("error", err.Error()).Debug("Handshake failed")
            remove <- peer.String()  // Notify main to remove this peer from its list
            return
        }
    }
    ctxLog.Debug("Handshake successful")

    // Cleanup
    defer func() {
        for i := range peer.workQueue {
            work <- peer.workQueue[i].index
        }
        peer.Conn.Close()  // Close the connection, results in the goroutine exiting due to error
        peer.Conn = nil  // Clear the connection for next time
        ctxLog.Debug("Peer shutdown")
    }()

    // Setup peer connection
    connection := make(chan []byte, 2)  // Buffer so that connection can exit if we haven't read the data yet
    go peer.Conn.Await(connection)
    peer.Conn.Timeout = peerTimeout

    // Work loop
    for {
        select {
        case data, ok := <-connection:
            if !ok {  // Connection failed
                remove <- peer.String()  // Notify main to remove this peer from its list
                return
            }
            msg := message.Decode(data)
            if err := peer.handleMessage(msg, work, results); err != nil {
                ctxLog.WithFields(log.Fields{"type": msg.String(), "size": len(msg.Payload), "error": err.Error()}).Debug("Error handling message")
                remove <- peer.String()  // Notify main to remove this peer from its list
                return
            }
        case _, ok := <-done:  // Check if the torrent told the peer to shutdown
            if !ok {
                return
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
                    remove <- peer.String()  // Notify main to remove this peer from its list
                    return
                }
            default:  // Don't block if we can't find work
            }
        }
    }
}
