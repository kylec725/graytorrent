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
)

// Errors
var (
	ErrPeerNotFound = errors.New("Peer not found")
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
	Path          string
	IncomingPeers chan peer.Peer     // Used by main to forward incoming peers
	Info          common.TorrentInfo // Contains meta data of the torrent
	Trackers      []tracker.Tracker
	Peers         []peer.Peer
}

// Setup gets and sets up necessary properties of a new torrent object
func (to *Torrent) Setup(ctx context.Context) error {
	port := common.Port(ctx)

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
	to.Trackers, err = tracker.GetTrackers(meta, port)
	if err != nil {
		return errors.Wrap(err, "Setup")
	}

	// Initialize files for writing
	if err := write.NewWrite(to.Info); err != nil {
		return errors.Wrap(err, "Setup")
	}

	// Make channel for incoming peers
	to.IncomingPeers = make(chan peer.Peer)

	return nil
}

// Start initiates a routine to download a torrent from peers
func (to *Torrent) Start(ctx context.Context) {
	log.WithField("name", to.Info.Name).Info("Torrent started")
	peers := make(chan peer.Peer)                     // For incoming peers from trackers
	work := make(chan int, to.Info.TotalPieces)       // Piece indices we need
	results := make(chan int, to.Info.TotalPieces)    // Notification that a piece is done
	remove := make(chan string)                       // For peers to notify they should be removed from our list
	complete := make(chan bool)                       // To notify trackers to send the completed message
	unchokeTicker := time.NewTicker(10 * time.Second) // Change who is unchoked after a period of time
	ctx, cancel := context.WithCancel(context.WithValue(ctx, common.KeyInfo, &to.Info))

	// Cleanup
	defer func() {
		to.Peers = nil // Clear peers
		cancel()       // Close all trackers and peers if the torrent goroutine returns
	}()

	// Start tracker goroutines
	for i := range to.Trackers {
		go to.Trackers[i].Run(ctx, peers, complete)
	}

	// Populate work queue
	for i := 0; i < to.Info.TotalPieces; i++ {
		if !to.Info.Bitfield.Has(i) {
			work <- i
		}
	}

	pieces := 0 // Counter of finished pieces
	for {
		select {
		case newPeer := <-peers: // Peers from trackers
			if !to.hasPeer(newPeer) {
				to.Peers = append(to.Peers, newPeer)
				go newPeer.StartWork(ctx, work, results, remove)
			}
		case newPeer := <-to.IncomingPeers: // Incoming peers from main
			if !to.hasPeer(newPeer) {
				to.Peers = append(to.Peers, newPeer)
				go newPeer.StartWork(ctx, work, results, remove)
			}
		case deadPeer := <-remove: // Don't exit as trackers may find peers
			to.removePeer(deadPeer)
		case index := <-results: // TODO change states
			to.Info.Bitfield.Set(index)
			to.Info.Left -= common.PieceSize(to.Info, index)
			go to.sendHave(index)
			pieces++
			if pieces == to.Info.TotalPieces {
				log.WithField("name", to.Info.Name).Info("Torrent completed")
				close(complete) // Notify trackers to send completed message
			}
		case <-unchokeTicker.C:
			if len(to.Peers) > 0 {
				to.unchokeAlg()
			}
		case <-ctx.Done():
			log.WithField("name", to.Info.Name).Info("Torrent stopped")
			return
		}
	}
}

func (to *Torrent) sendHave(index int) {
	msg := message.Have(uint32(index))
	for _, peer := range to.Peers {
		peer.SendMessage(msg)
	}
}

func (to *Torrent) removePeer(name string) {
	removeIndex := -1
	for i, peer := range to.Peers {
		if name == peer.String() {
			removeIndex = i
		}
	}
	if removeIndex == -1 {
		return
	}
	to.Peers[removeIndex] = to.Peers[len(to.Peers)-1]
	to.Peers = to.Peers[:len(to.Peers)-1]
}

func (to *Torrent) hasPeer(newPeer peer.Peer) bool {
	for _, peer := range to.Peers {
		if newPeer.Addr == peer.Addr {
			return true
		}
	}
	return false
}
