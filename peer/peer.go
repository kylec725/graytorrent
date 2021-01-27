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
    "context"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/bitfield"
    "github.com/kylec725/graytorrent/peer/message"
    "github.com/kylec725/graytorrent/connect"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const handshakeTimeout = 20 * time.Second
const peerTimeout = 120 * time.Second
const pollTimeout = 5 * time.Second
const startRate = 2  // Uses adaptive rate after first requests

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Addr string
    Conn *connect.Conn  // nil if not connected
    // Info *common.TorrentInfo

    bitfield bitfield.Bitfield
    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool
    rate int  // max number of outgoing requests/pieces a peer can queue
    workQueue []workPiece
    send chan message.Message
    shutdown chan bool
}

func (peer Peer) String() string {
    return peer.Addr
}

// Sends and receives a handshake from the peer
// func initHandshake() error {
//     if err := peer.SendHandshake(); err != nil {
//         return errors.Wrap(err, "initHandshake")
//     }
//     infoHash, err := peer.RcvHandshake()
//     if err != nil {
//         return errors.Wrap(err, "initHandshake")
//     } else if !bytes.Equal(peer.Info.InfoHash[:], infoHash[:]) {  // Verify the infohash
//         return errors.Wrap(ErrInfoHash, "initHandshake")
//     }
//     // Send bitfield to the peer
//     msg := message.Bitfield(peer.Info.Bitfield)
//     err = peer.Conn.Write(msg.Encode())
//     return errors.Wrap(err, "initHandshake")
// }

// New returns a new instantiated peer
func New(ctx context.Context, addr string, conn net.Conn) Peer {
    info := common.Info(ctx)
    var peerConn *connect.Conn = nil
    if conn != nil {
        peerConn = &connect.Conn{Conn: conn, Timeout: handshakeTimeout}
    }
    bitfieldSize := int(math.Ceil(float64(info.TotalPieces) / 8))
    return Peer{
        Addr: addr,
        Conn: peerConn,
        // Info: info,

        bitfield: make([]byte, bitfieldSize),
        amChoking: true,
        amInterested: false,
        peerChoking: true,
        peerInterested: false,
        rate: startRate,
        workQueue: []workPiece{},
        send: make(chan message.Message),
        shutdown: make(chan bool),
    }
}

// Send allows outside goroutines to send messages to a peer, not used internally
func (peer *Peer) Send(msg message.Message) {
    peer.send <- msg
}

func (peer *Peer) handleSend(msg message.Message) error {
    switch msg.ID {
    case message.MsgChoke:
        peer.amChoking = true
    case message.MsgUnchoke:
        peer.amChoking = false
    }
    err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "handleSend")
}

// Shutdown stops a Peer's work process
func (peer *Peer) Shutdown() {
    peer.shutdown <- true
}

// StartWork makes a peer wait for pieces to download
func (peer *Peer) StartWork(work chan int, results chan int, remove chan string) {
    peerLog := log.WithField("peer", peer.String())
    if peer.Conn == nil {
        err := peer.initHandshake()
        if err != nil {
            peerLog.WithField("error", err.Error()).Debug("Handshake failed")
            remove <- peer.String()  // Notify main to remove this peer from its list
            return
        }
    }
    peerLog.Debug("Handshake successful")

    // Cleanup
    defer func() {
        for i := range peer.workQueue {
            work <- peer.workQueue[i].index
        }
        peer.Conn.Close()  // Close the connection, results in the goroutine exiting due to error
        peerLog.Debug("Peer shutdown")
    }()

    // Setup peer connection
    connection := make(chan []byte, 2)  // Buffer so that connection can exit if we haven't read the data yet
    go peer.Conn.Await(connection)
    peer.Conn.Timeout = peerTimeout

    // Work loop
    for {
        select {
        case data, ok := <-connection:  // Incoming data from peer
            if !ok {  // Connection failed
                remove <- peer.String()  // Notify main to remove this peer from its list
                return
            }
            msg := message.Decode(data)
            if err := peer.handleMessage(msg, work, results); err != nil {
                peerLog.WithFields(log.Fields{"type": msg.String(), "size": len(msg.Payload), "error": err.Error()}).Debug("Error handling message")
                remove <- peer.String()  // Notify main to remove this peer from its list
                return
            }
        case msg := <-peer.send:
            if err := peer.handleSend(msg); err != nil {
                peerLog.WithFields(log.Fields{"type": msg.String(), "error": err.Error()}).Debug("Error sending message")
                remove <- peer.String()
            }
        case <-peer.shutdown:  // Check if the torrent told the peer to shutdown
            return
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
                    peerLog.WithFields(log.Fields{"piece index": index, "error": err.Error()}).Debug("Failed to start piece download")
                    work <- index  // Put piece back onto work channel
                    remove <- peer.String()
                    return
                }
            default:  // Don't block if we can't find work
            }
        }
    }
}
