/*
Package torrent provides a library for reading from a torrent file
and storing desired information for leeching or seeding.
Tracker file to handle grabbing information about current
peers and the state of the file.
Write file to handle writing and getting pieces, as well as verifying
the hash of received pieces.
Will communicate with the peers package for sending and receiving
pieces of the torrent.
*/
package torrent

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/config"
	"github.com/kylec725/graytorrent/internal/peer"
	"github.com/kylec725/graytorrent/internal/peer/message"
	"github.com/kylec725/graytorrent/internal/tracker"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Errors
var (
	ErrPeerNotFound = errors.New("Peer not found")

	currID uint32 = 1 // 0 is reserved for an unset ID
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
	ID       uint32              `json:"-"`      // Can be used by grpc client to select a torrent
	File     string              `json:"File"`   // .torrent file
	Magnet   string              `json:"Magnet"` // Magnet link
	Info     *common.TorrentInfo `json:"Info"`   // Contains meta data of the torrent
	Trackers []*tracker.Tracker  `json:"Trackers"`
	Peers    []*peer.Peer        `json:"-"`
	NewPeers chan peer.Peer      `json:"-"` // Used by main and trackers to send in new peers
	Started  bool                `json:"-"` // Flag to see if torrent goroutine is running

	cancel            context.CancelFunc `json:"-"` // Cancel function for context, we can use it to see if the Start goroutine is running
	optimisticUnchoke *peer.Peer         `json:"-"` // The peer that is currently optimistically unchoked
}

// Init initializes a torrent so that it is ready to download or seed
func (to *Torrent) Init() error {
	// TODO: use either file or magnet link
	if to.Info == nil {
		var err error
		// Convert to a TorrentInfo struct
		to.Info, to.Trackers, err = InfoFromFile(to.File)
		if err != nil {
			return errors.Wrap(err, "Init")
		}
	}

	to.NewPeers = make(chan peer.Peer)

	to.Started = false

	_, to.cancel = context.WithCancel(context.Background()) // Dummy function so that stopping a torrent should always succeed

	to.ID = currID
	currID++

	return nil
}

// Start initiates a routine to download a torrent from peers
func (to *Torrent) Start(ctx context.Context) {
	to.Started = true
	torrentLog := log.WithFields(log.Fields{"name": to.Info.Name, "infohash": hex.EncodeToString(to.Info.InfoHash[:])})
	torrentLog.Info("Torrent started")
	work := make(chan int, to.Info.TotalPieces)       // Piece indices we need
	results := make(chan int, to.Info.TotalPieces)    // Notification that a piece is done
	complete := make(chan bool)                       // Notify trackers that the torrent is complete
	deadPeers := make(chan string)                    // For peers to notify they should be removed from our list
	unchokeTicker := time.NewTicker(10 * time.Second) // Change who is unchoked after a period of time
	lastOpUnchoke := time.Now()                       // Keep track of when the optimistic unchoke was changed
	ctx, cancel := context.WithCancel(ctx)
	to.cancel = cancel

	// Cleanup
	defer func() {
		to.Started = false
		unchokeTicker.Stop()
		to.Peers = nil // Clear peers
		cancel()       // Close all trackers and peers if the torrent goroutine returns
		to.optimisticUnchoke = nil
	}()

	// Start tracker goroutines
	for i := range to.Trackers {
		go to.Trackers[i].Run(ctx, to.Info, to.NewPeers, complete)
	}

	// Populate work queue
	for i := 0; i < to.Info.TotalPieces; i++ { // TODO: change to random order or a priority queue (use heap)
		if !to.Info.Bitfield.Has(i) {
			work <- i
		}
	}

	for {
		select {
		case <-ctx.Done(): // TODO: use a waitgroup to make sure trackers and peers properly close out
			torrentLog.Info("Torrent stopped")
			return
		case deadPeer := <-deadPeers: // Don't exit since trackers may find peers
			to.removePeer(deadPeer)
		case newPeer := <-to.NewPeers: // Incoming peers that contacted us
			if !to.hasPeer(newPeer.String()) && len(to.Peers) < config.GetConfig().Network.MaxTorrentConnections {
				go to.addPeer(ctx, &newPeer, work, results, deadPeers)
			}
		case index := <-results:
			to.Info.Bitfield.Set(index)
			to.Info.Left -= to.Info.PieceSize(index)
			msg := message.Have(uint32(index)) // Notify peers that we have a new piece
			for i := range to.Peers {
				to.Peers[i].Send <- msg
			}

			if to.Info.Left == 0 { // WARNING: goroutine leak when seeding begins (high cpu usage)
				torrentLog.Info("Torrent completed")
				close(complete)
			}
		case <-unchokeTicker.C:
			if len(to.Peers) > 0 {
				if time.Since(lastOpUnchoke) > 30*time.Second || to.optimisticUnchoke == nil {
					to.changeOptimisticUnchoke(&lastOpUnchoke)
				}
				to.unchokeAlg()
			}
		}
	}
}

// Stop stops the download or upload of a torrent
func (to *Torrent) Stop() {
	to.cancel()
}

func (to *Torrent) addPeer(ctx context.Context, p *peer.Peer, work chan int, results chan int, deadPeers chan string) {
	if p.Conn == nil {
		if err := p.Dial(); err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "peer": p.String()}).Debug("Dial failed")
			return
		} else if err := p.InitHandshake(to.Info); err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "peer": p.String()}).Debug("Handshake failed")
			return
		}
	}
	log.WithField("peer", p.String()).Debug("Handshake successful")
	to.Peers = append(to.Peers, p)
	p.StartWork(ctx, to.Info, work, results, deadPeers)
}

func (to *Torrent) removePeer(p string) {
	for i := range to.Peers {
		if to.Peers[i].String() == p {
			to.Peers[i] = to.Peers[len(to.Peers)-1]
			to.Peers = to.Peers[:len(to.Peers)-1]
			return
		}
	}
}

func (to *Torrent) hasPeer(p string) bool {
	for i := range to.Peers {
		if to.Peers[i].String() == p {
			return true
		}
	}
	return false
}

// DownRate returns the current total download rate of the torrent in bytes/sec
func (to *Torrent) DownRate() uint32 {
	totalRate := uint32(0)
	for i := range to.Peers {
		totalRate += to.Peers[i].DownRate()
	}
	return totalRate
}

// UpRate returns the current total download rate of the torrent in bytes/sec
func (to *Torrent) UpRate() uint32 {
	totalRate := uint32(0)
	for i := range to.Peers {
		totalRate += to.Peers[i].UpRate()
	}
	return totalRate
}
