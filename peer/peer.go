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
const pollTimeout = 5 * time.Second  // So that connection loops don't run too fast or get unstuck to check for more work
const keepAlive = 120 * time.Second  // How long to wait before removing a peer with no messages
const startRate = 2  // Uses adaptive rate after first requests

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    Addr string
    Conn *connect.Conn  // nil if not connected

    bitfield bitfield.Bitfield
    amChoking bool
    amInterested bool
    peerChoking bool
    peerInterested bool
    rate int  // max number of outgoing requests/pieces a peer can queue
    workQueue []workPiece
    lastContact time.Time
    send chan message.Message
    shutdown chan bool
}

func (peer Peer) String() string {
    return peer.Addr
}

// New returns a new instantiated peer
func New(addr string, conn net.Conn, info common.TorrentInfo) Peer {
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

// SendMessage allows outside goroutines to send messages to a peer, not used internally
func (peer *Peer) SendMessage(msg message.Message) {
    peer.send <- msg
}

func (peer *Peer) handleSend(msg message.Message) error {
    switch msg.ID {
    case message.MsgChoke:
        peer.amChoking = true
    case message.MsgUnchoke:
        peer.amChoking = false
    }
    _, err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "handleSend")
}

// Shutdown stops a Peer's work process
func (peer *Peer) Shutdown() {
    peer.shutdown <- true
}

// StartWork makes a peer wait for pieces to download
func (peer *Peer) StartWork(ctx context.Context, work chan int, results chan int, remove chan string) {
    info := common.Info(ctx)
    peerLog := log.WithField("peer", peer.String())
    if peer.Conn == nil {
        if err := peer.dial(); err != nil {
            peerLog.WithField("error", err.Error()).Debug("Dial failed")
            remove <- peer.String()  // Notify main to remove this peer from its list
            return
        } else if err := peer.initHandshake(info); err != nil {
            peerLog.WithField("error", err.Error()).Debug("Handshake failed")
            remove <- peer.String()
            return
        }
    }
    peerLog.Debug("Handshake successful")

    // Setup peer connection
    connCtx, connCancel := context.WithCancel(ctx)
    peer.Conn.Timeout = pollTimeout
    connection := make(chan []byte, 2)  // Buffer so that connection can exit if we haven't read the data yet
    go peer.Conn.Poll(connCtx, connection)

    // Cleanup
    defer func() {
        for i := range peer.workQueue {
            work <- peer.workQueue[i].index
        }
        connCancel()
        peerLog.Debug("Peer shutdown")
    }()

    // Work loop
    for {
        select {
        case data, ok := <-connection:  // Incoming data from peer
            if !ok {  // Connection failed
                remove <- peer.String()  // Notify main to remove this peer from its list
                return
            }
            peer.lastContact = time.Now()
            currInfo := common.Info(ctx)
            msg := message.Decode(data)
            if err := peer.handleMessage(msg, currInfo, work, results); err != nil {
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
            if time.Since(peer.lastContact) >= keepAlive {  // Check if peer has passed the keep-alive time
                remove <- peer.String()
                return
            }
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
                err := peer.downloadPiece(info, index)
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
