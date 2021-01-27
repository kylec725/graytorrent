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
    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/metainfo"
    "github.com/kylec725/graytorrent/tracker"
    "github.com/kylec725/graytorrent/peer"
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
    Path string
    Port uint16
    IncomingPeers chan peer.Peer  // Used by main to forward incoming peers
    Info common.TorrentInfo  // Contains meta data of the torrent
    Trackers []tracker.Tracker
    Peers []peer.Peer

    stop chan bool
}

// Setup gets and sets up necessary properties of a new torrent object
func (to *Torrent) Setup() error {
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
    to.Trackers, err = tracker.GetTrackers(meta, &to.Info, to.Port)
    if err != nil {
        return errors.Wrap(err, "Setup")
    }

    // Initialize files for writing
    if err := write.NewWrite(&to.Info); err != nil {
        return errors.Wrap(err, "Setup")
    }

    // Make the shutdown channel
    to.stop = make(chan bool)

    return nil
}

// Stop signals a torrent to stop downloading
func (to *Torrent) Stop() {
    to.stop <- true
}

// Start initiates a routine to download a torrent from peers
func (to *Torrent) Start() {
    log.WithField("name", to.Info.Name).Info("Torrent started")
    peers := make(chan peer.Peer)                  // For incoming peers from trackers
    work := make(chan int, to.Info.TotalPieces)    // Piece indices we need
    results := make(chan int, to.Info.TotalPieces) // Notification that a piece is done
    remove := make(chan string)                    // For peers to notify they should be removed from our list

    // Cleanup
    defer func() {
        for _, peer := range to.Peers {
            peer.Shutdown()
        }
        for _, tracker := range to.Trackers {
            tracker.Shutdown(to.Info.Left)
        }
        to.Peers = nil  // Clear peers
    }()

    // Start tracker goroutines
    for i := range to.Trackers {
        go to.Trackers[i].Run(to.Info.Left, peers)
    }

    // Populate work queue
    for i := 0; i < to.Info.TotalPieces; i++ {
        if !to.Info.Bitfield.Has(i) {
            work <- i
        }
    }

    pieces := 0  // Counter of finished pieces
    for {
        select {
        case newPeer := <-peers:  // Peers from trackers
            to.Peers = append(to.Peers, newPeer)
            go newPeer.StartWork(work, results, remove)
        case newPeer := <-to.IncomingPeers:  // Incoming peers from main
            to.Peers = append(to.Peers, newPeer)
            go newPeer.StartWork(work, results, remove)
        case deadPeer := <-remove:
            to.removePeer(deadPeer)
            if len(to.Peers) == 0 {  // Exit if we don't have anymore peers
                return
            }
        case index := <-results:  // TODO change states
            to.Info.Bitfield.Set(index)
            to.Info.Left -= common.PieceSize(&to.Info, index)
            go to.sendHave(index)
            pieces++
            if pieces == to.Info.TotalPieces {
                log.WithField("name", to.Info.Name).Info("Torrent completed")
            }
        case <-to.stop:
            log.WithField("name", to.Info.Name).Info("Torrent stopped")
            return
        }
    }
}

// Save saves data about a managed torrent's state to a file
func (to *Torrent) Save() {
    // TODO log results of saving
    // TODO consider have a directory, with a file for each torrent's state
    // TODO alternative: open history file json maybe, see if we are in it, if not: add ourselves
    //      if we are already, update info
    return
}

func (to *Torrent) sendHave(index int) {
    var err error
    for _, peer :=  range to.Peers {
        if err = peer.Have(index); err != nil {
            log.WithFields(log.Fields{"peer": peer.String(), "error": err.Error()}).Debug("Error sending have message")
        }
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
    // Notify the peer to shutdown if it hasn't already
    to.Peers[removeIndex] = to.Peers[len(to.Peers) - 1]
    to.Peers = to.Peers[:len(to.Peers) - 1]
}
