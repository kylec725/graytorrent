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
    Info common.TorrentInfo
    Trackers []tracker.Tracker
    Peers []peer.Peer
    Port uint16
    IncomingPeers chan peer.Peer  // Used by main to forward incoming peers

    shutdown bool
    // TODO save path, left, bitfield, peerid somewhere to keep track of state
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

    return nil
}

// Stop signals a torrent to stop downloading
func (to *Torrent) Stop() {
    to.shutdown = true
    log.WithField("name", to.Info.Name).Info("Torrent stopped")
}

// Start initiates a routine to download a torrent from peers
func (to *Torrent) Start() {
    log.WithField("name", to.Info.Name).Info("Torrent started")
    to.shutdown = false
    peers := make(chan peer.Peer)                  // For incoming peers from trackers
    work := make(chan int, to.Info.TotalPieces)    // Piece indices we need
    results := make(chan int, to.Info.TotalPieces) // Notification that a piece is done
    remove := make(chan string)                    // For peers to notify they should be removed from our list
    done := make(chan bool)                        // Notify goroutines to quit

    // Start tracker goroutines
    for i := range to.Trackers {
        go to.Trackers[i].Run(peers, done)
    }

    // Populate work queue
    for i := 0; i < to.Info.TotalPieces; i++ {
        if !to.Info.Bitfield.Has(i) {
            work <- i
        }
    }

    pieces := 0  // Counter of finished pieces
    for {
        if to.shutdown {
            goto exit
        }

        select {
        case newPeer := <-peers:  // peers from trackers
            to.Peers = append(to.Peers, newPeer)
            go newPeer.StartWork(work, results, remove, done)
        case newPeer := <-to.IncomingPeers:  // incoming peers from main
            to.Peers = append(to.Peers, newPeer)
            go newPeer.StartWork(work, results, remove, done)
        case deadPeer := <-remove:
            to.removePeer(deadPeer)
        case index := <-results:
            to.Info.Bitfield.Set(index)
            pieces++
            if pieces == to.Info.TotalPieces {
                // TODO go to seeding mode after finishing the download
                log.WithField("name", to.Info.Name).Info("Torrent finished")
                goto exit
            }
        }
    }

    exit:
    close(done)
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
    to.Peers[removeIndex].Shutdown()
    to.Peers[removeIndex] = to.Peers[len(to.Peers) - 1]
    to.Peers = to.Peers[:len(to.Peers) - 1]
}
