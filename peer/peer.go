/*
Package peer provides the ability to setup connections with peers as well
as manage sending and receiving torrent pieces with those peers.
Peers also handle writing pieces to file if necessary.
*/
package peer

import (
	"context"
	"math"
	"net"
	"time"

	"github.com/kylec725/graytorrent/bitfield"
	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/connect"
	"github.com/kylec725/graytorrent/peer/message"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const peerTimeout = 20 * time.Second    // Time to wait on an expected peer connection operation
const workTimeout = 5 * time.Second     // To get unstuck if we need to get work
const requestTimeout = 30 * time.Second // How long to wait on requests before sending work back
const keepAlive = 120 * time.Second     // How long to wait before removing a peer with no messages
const startRate = 2                     // Uses adaptive rate after first requests

// Peer stores info about connecting to peers as well as their state
type Peer struct {
	Addr string
	Conn *connect.Conn // nil if not connected

	bitfield       bitfield.Bitfield
	amChoking      bool
	amInterested   bool
	peerChoking    bool
	peerInterested bool
	rate           int // max number of outgoing requests/pieces a peer can queue
	workQueue      []workPiece
	lastContact    time.Time
	lastRequest    time.Time
	send           chan message.Message // Used for torrent goroutine to send messages
}

func (p Peer) String() string {
	return p.Addr
}

// New returns a new instantiated peer
func New(addr string, conn net.Conn, info common.TorrentInfo) Peer {
	var peerConn *connect.Conn = nil
	if conn != nil {
		peerConn = &connect.Conn{Conn: conn, Timeout: peerTimeout}
	}
	bitfieldSize := int(math.Ceil(float64(info.TotalPieces) / 8))
	return Peer{
		Addr: addr,
		Conn: peerConn,

		bitfield:       make([]byte, bitfieldSize),
		amChoking:      true,
		amInterested:   false,
		peerChoking:    true,
		peerInterested: false,
		rate:           startRate,
		lastContact:    time.Now(),
		lastRequest:    time.Now(),
		workQueue:      []workPiece{},
		send:           make(chan message.Message),
	}
}

// SendMessage allows outside goroutines to send messages to a peer, not used internally
func (p *Peer) SendMessage(msg message.Message) {
	p.send <- msg
}

func (p *Peer) handleSend(msg message.Message) error {
	switch msg.ID {
	case message.MsgChoke:
		p.amChoking = true
	case message.MsgUnchoke:
		p.amChoking = false
	}
	_, err := p.Conn.Write(msg.Encode())
	return errors.Wrap(err, "handleSend")
}

// StartWork makes a peer wait for pieces to download
func (p *Peer) StartWork(ctx context.Context, work chan int, results chan int, remove chan string) {
	info := common.Info(ctx)
	peerLog := log.WithField("peer", p.String())
	if p.Conn == nil {
		if err := p.dial(); err != nil {
			peerLog.WithField("error", err.Error()).Debug("Dial failed")
			remove <- p.String() // Notify main to remove this peer from its list
			return
		} else if err := p.initHandshake(info); err != nil {
			peerLog.WithField("error", err.Error()).Debug("Handshake failed")
			remove <- p.String()
			return
		}
	}
	peerLog.Debug("Handshake successful")

	// Setup peer connection
	connCtx, connCancel := context.WithCancel(ctx)
	p.Conn.Timeout = peerTimeout
	connection := make(chan []byte, 2) // Buffer so that connection can exit if we haven't read the data yet
	go p.Conn.Poll(connCtx, connection)

	// Cleanup
	defer func() {
		remove <- p.String() // Notify main to remove this peer from its list
		p.clearWork(work)
		connCancel()
		peerLog.Debug("Peer shutdown")
	}()

	// Work loop
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-connection: // Incoming data from peer
			if !ok { // Connection failed
				return
			}
			p.lastContact = time.Now()
			currInfo := common.Info(ctx)
			msg := message.Decode(data)
			if err := p.handleMessage(msg, currInfo, work, results); err != nil {
				peerLog.WithFields(log.Fields{"type": msg.String(), "size": len(msg.Payload), "error": err.Error()}).Debug("Error handling message")
				return
			}
		case msg := <-p.send:
			if err := p.handleSend(msg); err != nil {
				peerLog.WithFields(log.Fields{"type": msg.String(), "error": err.Error()}).Debug("Error sending message")
				return
			}
		case <-time.After(workTimeout): // Poll to get unstuck if no messages are received
			if time.Since(p.lastRequest) >= requestTimeout {
				p.clearWork(work)
			}
			if time.Since(p.lastContact) >= keepAlive { // Check if peer has passed the keep-alive time
				return
			}
		}

		// Find new work piece if queue is open
		if len(p.workQueue) < p.rate {
			select {
			case index := <-work:
				// Send the work back if the peer does not have the piece
				if !p.bitfield.Has(index) {
					work <- index
					continue
				}

				// Download piece from the peer
				err := p.downloadPiece(info, index)
				if err != nil {
					peerLog.WithFields(log.Fields{"piece index": index, "error": err.Error()}).Debug("Failed to start piece download")
					return
				}
			default: // Don't block if we can't find work
			}
		}
	}
}
