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
	log "github.com/sirupsen/logrus"
)

const peerTimeout = 20 * time.Second       // Time to wait on an expected peer connection operation
const workTimeout = 5 * time.Second        // To get unstuck if we need to get work
const requestTimeout = 30 * time.Second    // How long to wait on requests before sending work back
const receiveKeepAlive = 120 * time.Second // How long to wait before removing a peer with no messages
const sendKeepAlive = 90 * time.Second     // How long to wait before sending a keep alive message
const startRate = 2                        // Uses adaptive rate after first requests
const adjustTime = 5                       // How often in seconds to adjust the queuing rate

// Peer stores info about connecting to peers as well as their state
type Peer struct {
	Addr           string
	Conn           *connect.Conn // nil if not connected
	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
	Rate           int // Peer's download rate in KiB/s

	bitfield            bitfield.Bitfield
	workQueue           []workPiece
	kbReceived          int // Number of kilobytes of data received since the last adjustment time
	lastMessageReceived time.Time
	lastMessageSent     time.Time
	lastRequest         time.Time
	send                chan message.Message // Used for torrent goroutine to send messages
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
		Addr:           addr,
		Conn:           peerConn,
		AmChoking:      true,
		AmInterested:   false,
		PeerChoking:    true,
		PeerInterested: false,
		Rate:           startRate,

		bitfield:            make([]byte, bitfieldSize),
		workQueue:           []workPiece{},
		kbReceived:          0,
		lastMessageReceived: time.Now(),
		lastMessageSent:     time.Now(),
		lastRequest:         time.Now(),
		send:                make(chan message.Message),
	}
}

// SendMessage allows outside goroutines to send messages to a peer, not used internally
func (p *Peer) SendMessage(msg message.Message) {
	p.send <- msg
}

// StartWork makes a peer wait for pieces to download
func (p *Peer) StartWork(ctx context.Context, work chan int, results chan int, deadPeers chan string) {
	info := common.Info(ctx)
	peerLog := log.WithField("peer", p.String())
	if p.Conn == nil {
		if err := p.dial(); err != nil {
			peerLog.WithField("error", err.Error()).Debug("Dial failed")
			deadPeers <- p.String() // Notify main to remove this peer from its list
			return
		} else if err := p.initHandshake(info); err != nil {
			peerLog.WithField("error", err.Error()).Debug("Handshake failed")
			deadPeers <- p.String()
			return
		}
	}
	peerLog.Debug("Handshake successful")

	// Setup peer connection
	connCtx, connCancel := context.WithCancel(ctx)
	p.Conn.Timeout = peerTimeout
	connection := make(chan []byte, 2) // Buffer so that connection can exit if we haven't read the data yet
	go p.Conn.Poll(connCtx, connection)

	// Create ticker to adjust queuing rate
	rateTicker := time.NewTicker(adjustTime * time.Second)

	// Cleanup
	defer func() {
		deadPeers <- p.String() // Notify main to remove this peer from its list
		p.clearWork(work)
		connCancel()
		rateTicker.Stop()
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
			p.lastMessageReceived = time.Now()
			currInfo := common.Info(ctx)
			msg := message.Decode(data)
			if err := p.handleMessage(msg, currInfo, work, results); err != nil {
				peerLog.WithFields(log.Fields{"type": msg.String(), "size": len(msg.Payload), "error": err.Error()}).Debug("Error handling message")
				return
			}
		case msg := <-p.send:
			if err := p.sendMessage(&msg); err != nil {
				peerLog.WithFields(log.Fields{"type": msg.String(), "error": err.Error()}).Debug("Error sending message")
				return
			}
		case <-rateTicker.C:
			p.adjustRate()
			p.kbReceived = 0
		case <-time.After(workTimeout): // Poll to get unstuck if no messages are received
			if time.Since(p.lastRequest) >= requestTimeout {
				p.clearWork(work)
			}
			if time.Since(p.lastMessageSent) >= sendKeepAlive {
				p.sendMessage(nil)
			}
			if time.Since(p.lastMessageReceived) >= receiveKeepAlive { // Check if peer has passed the keep-alive time
				return
			}
		}

		// Find new work piece if queue is open
		if len(p.workQueue) < p.Rate {
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
