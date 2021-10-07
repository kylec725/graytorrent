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
	"time"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/metainfo"
	"github.com/kylec725/graytorrent/peer"
	"github.com/kylec725/graytorrent/peer/message"
	"github.com/kylec725/graytorrent/tracker"
	"github.com/kylec725/graytorrent/write"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

// Errors
var (
	ErrPeerNotFound  = errors.New("Peer not found")
	ErrTorrentExists = errors.New("Torrent is already being managed")
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
	Path      string             `json:"Path"`
	NewPeers  chan peer.Peer     `json:"-"`    // Used by main and trackers to send in new peers
	Info      common.TorrentInfo `json:"Info"` // Contains meta data of the torrent
	Trackers  []tracker.Tracker  `json:"Trackers"`
	Peers     []peer.Peer        `json:"-"`
	deadPeers []string           `json:"-"`
	Cancel    context.CancelFunc `json:"-"` // Cancel function for context, we can use it to see if the Start goroutine is running
}

// Setup gets and sets up necessary properties of a new torrent object
func (to *Torrent) Setup(ctx context.Context) error {
	// Get metainfo
	meta, err := metainfo.Meta(to.Path)
	if err != nil {
		return errors.Wrap(err, "Setup")
	}

	// Convert to a TorrentInfo struct
	to.Info, err = common.GetInfo(meta)
	if err != nil {
		return errors.Wrap(err, "Setup")
	}

	// Create trackers list from metainfo announce or announce-list
	to.Trackers, err = tracker.GetTrackers(meta)
	if err != nil {
		return errors.Wrap(err, "Setup")
	}

	// Initialize files for writing
	if err := write.NewWrite(to.Info); err != nil { // Should fail if torrent already is being managed
		return errors.Wrap(err, "Setup")
	}

	// Make channel for incoming peers
	to.NewPeers = make(chan peer.Peer)

	to.Cancel = nil

	return nil
}

// Start initiates a routine to download a torrent from peers
func (to *Torrent) Start(ctx context.Context) {
	log.WithField("name", to.Info.Name).Info("Torrent started")
	work := make(chan int, to.Info.TotalPieces)       // Piece indices we need
	results := make(chan int, to.Info.TotalPieces)    // Notification that a piece is done
	deadPeers := make(chan string)                    // For peers to notify they should be removed from our list
	complete := make(chan bool)                       // To notify trackers to send the completed message
	unchokeTicker := time.NewTicker(10 * time.Second) // Change who is unchoked after a period of time
	ctx, cancel := context.WithCancel(context.WithValue(ctx, common.KeyInfo, &to.Info))
	to.Cancel = cancel

	// Cleanup
	defer func() {
		unchokeTicker.Stop()
		to.Peers = nil     // Clear peers
		to.deadPeers = nil // Clear dead peers
		cancel()           // Close all trackers and peers if the torrent goroutine returns
		to.Cancel = nil    // Make cancel func nil so that state can see if the torrent was started
	}()

	// Start tracker goroutines
	for i := range to.Trackers {
		go to.Trackers[i].Run(ctx, to.NewPeers, complete)
	}

	// Populate work queue
	for i := 0; i < to.Info.TotalPieces; i++ { // TODO: change to random order or a priority queue (use heap)
		if !to.Info.Bitfield.Has(i) {
			work <- i
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.WithField("name", to.Info.Name).Info("Torrent stopped")
			return
		case deadPeer := <-deadPeers: // Don't exit since trackers may find peers
			to.removePeer(deadPeer)
		case newPeer := <-to.NewPeers: // Incoming peers from main
			if !to.hasPeer(newPeer) && len(to.Peers) < viper.GetInt("network.connections.torrentMax") {
				to.Peers = append(to.Peers, newPeer)
				go newPeer.StartWork(ctx, work, results, deadPeers)
			}
		case index := <-results:
			to.Info.Bitfield.Set(index)
			to.Info.Left -= common.PieceSize(to.Info, index)
			msg := message.Have(uint32(index)) // Notify peers that we have a new piece
			for i := range to.Peers {
				to.Peers[i].SendMessage(msg)
			}

			if to.Info.Left == 0 {
				log.WithField("name", to.Info.Name).Info("Torrent completed")
				close(complete) // Notify trackers to send completed message
			}
		case <-unchokeTicker.C:
			if len(to.Peers) > 0 {
				to.unchokeAlg()
			}
		}
	}
}

func (to *Torrent) removePeer(name string) {
	for i := range to.Peers {
		if name == to.Peers[i].String() {
			to.Peers[i] = to.Peers[len(to.Peers)-1]
			to.Peers = to.Peers[:len(to.Peers)-1]
			return
		}
	}
}

func (to *Torrent) hasPeer(peer peer.Peer) bool {
	for i := range to.Peers {
		if peer.Addr == to.Peers[i].Addr {
			return true
		}
	}
	return false
}

// Rate returns the current total download rate of the torrent in kb/sec
func (to *Torrent) Rate() int {
	totalRate := 0
	for i := range to.Peers {
		totalRate += to.Peers[i].Rate()
	}
	return totalRate
}
