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
	Path          string             `json:"Path"`
	IncomingPeers chan peer.Peer     `json:"-"`    // Used by main to forward incoming peers
	Info          common.TorrentInfo `json:"Info"` // Contains meta data of the torrent
	Trackers      []tracker.Tracker  `json:"Trackers"`
	Peers         []peer.Peer        `json:"-"`
	deadPeers     []string           `json:"-"`
	Ctx           context.Context    `json:"-"`
	Cancel        context.CancelFunc `json:"-"`
	State         State              `json:"-"`
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

	to.Ctx = ctx

	to.State = Stopped
	// TODO: set state as Complete based off save data

	return nil
}

// Start initiates a routine to download a torrent from peers
func (to *Torrent) Start() {
	log.WithField("name", to.Info.Name).Info("Torrent started")
	peers := make(chan peer.Peer)                     // For incoming peers from trackers
	work := make(chan int, to.Info.TotalPieces)       // Piece indices we need
	results := make(chan int, to.Info.TotalPieces)    // Notification that a piece is done
	remove := make(chan string)                       // For peers to notify they should be removed from our list
	complete := make(chan bool)                       // To notify trackers to send the completed message
	unchokeTicker := time.NewTicker(10 * time.Second) // Change who is unchoked after a period of time
	ctx, cancel := context.WithCancel(context.WithValue(to.Ctx, common.KeyInfo, &to.Info))
	to.Cancel = cancel

	// Change state
	if to.State == Stopped {
		to.State = Started
	} else if to.State == Complete {
		to.State = Seeding
	}

	// Cleanup
	defer func() {
		to.Peers = nil     // Clear peers
		to.deadPeers = nil // Clear dead peers
		cancel()           // Close all trackers and peers if the torrent goroutine returns
		if to.State == Started {
			to.State = Stopped
		} else if to.State == Seeding {
			to.State = Complete
		}
	}()

	// Start tracker goroutines
	for i := range to.Trackers {
		go to.Trackers[i].Run(ctx, peers, complete)
	}

	// Populate work queue
	for i := 0; i < to.Info.TotalPieces; i++ { // TODO: change to random order
		if !to.Info.Bitfield.Has(i) {
			work <- i
		}
	}

	pieces := 0 // Counter of finished pieces
	for {
		select {
		case deadPeer := <-remove: // Don't exit as trackers may find peers
			to.removePeer(deadPeer)
		case newPeer := <-peers: // Peers from trackers
			if !to.hasPeer(newPeer) && !to.isDeadPeer(newPeer) {
				to.Peers = append(to.Peers, newPeer)
				go newPeer.StartWork(ctx, work, results, remove)
			}
		case newPeer := <-to.IncomingPeers: // Incoming peers from main
			if !to.hasPeer(newPeer) {
				if to.isDeadPeer(newPeer) {
					to.removeDeadPeer(newPeer.Addr)
				}
				to.Peers = append(to.Peers, newPeer)
				go newPeer.StartWork(ctx, work, results, remove)
			}
		case index := <-results:
			to.Info.Bitfield.Set(index)
			to.Info.Left -= common.PieceSize(to.Info, index)
			go to.sendHave(index)
			pieces++
			if pieces == to.Info.TotalPieces {
				log.WithField("name", to.Info.Name).Info("Torrent completed")
				close(complete) // Notify trackers to send completed message
				to.State = Seeding
			}
		case <-unchokeTicker.C:
			if len(to.Peers) > 0 {
				to.unchokeAlg()
			}
		case <-ctx.Done():
			log.WithField("name", to.Info.Name).Info("Torrent stopped")
			return
		}
		// Adjust download state based off number of peers
		if to.State == Started && len(to.Peers) == 0 {
			to.State = Stalled
		} else if to.State == Stalled && len(to.Peers) > 0 {
			to.State = Started
		}
	}
}

func (to *Torrent) sendHave(index int) {
	msg := message.Have(uint32(index))
	for i := range to.Peers {
		to.Peers[i].SendMessage(msg)
	}
}

// TODO: change to move dead peers into a separate list so that we do not recontact, allow them to init connection with us though
func (to *Torrent) removePeer(name string) {
	for i := range to.Peers {
		if name == to.Peers[i].String() {
			to.deadPeers = append(to.deadPeers, to.Peers[i].String())

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

func (to *Torrent) removeDeadPeer(addr string) {
	for i := range to.deadPeers {
		if addr == to.deadPeers[i] {
			to.deadPeers[i] = to.deadPeers[len(to.deadPeers)-1]
			to.deadPeers = to.deadPeers[:len(to.deadPeers)-1]
			return
		}
	}
}

func (to *Torrent) isDeadPeer(peer peer.Peer) bool {
	for i := range to.deadPeers {
		if peer.Addr == to.deadPeers[i] {
			return true
		}
	}
	return false
}
