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
    "github.com/kylec725/graytorrent/peer"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrPeerNotFound = errors.New("Peer not found")
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
    Path string
    Info common.TorrentInfo
    Trackers []Tracker
    Peers []peer.Peer
    shutdown bool
}

// Setup gets and sets up necessary properties of a new torrent object
func (to *Torrent) Setup() error {
    // Get metainfo
    meta, err := metainfo.Meta(to.Path)
    if err != nil {
        return errors.Wrap(err, "Setup")
    }

    to.Info, err = common.GetInfo(meta)
    if err != nil {
        return errors.Wrap(err, "Setup")
    }

    // Create trackers list from metainfo announce or announce-list
    to.Trackers, err = getTrackers(meta, &to.Info)
    if err != nil {
        return errors.Wrap(err, "Setup")
    }

    return nil
}

func (to *Torrent) removePeer(name string) error {
    removeIndex := -1
    for i := range to.Peers {
        if name == to.Peers[i].String() {
            removeIndex = i
        }
    }
    if removeIndex == -1 {
        return errors.Wrap(ErrPeerNotFound, "removePeer")
    }
    // Close the connection before returning
    if err := to.Peers[removeIndex].Conn.Close(); err != nil {
        return errors.Wrap(err, "removePeer")
    }
    to.Peers[removeIndex] = to.Peers[len(to.Peers) - 1]
    to.Peers = to.Peers[:len(to.Peers) - 1]
    return nil
}

// Shutdown lets main signal a torrent to stop downloading
func (to *Torrent) Shutdown() {
    to.shutdown = true
}

// Download starts a routine to download a torrent from peers
// TODO
func (to *Torrent) Download() {
    to.shutdown = false
    peers := make(chan peer.Peer) // For incoming peers from trackers
    work := make(chan int)        // Piece indices we need
    remove := make(chan string)   // For peers to notify they should be removed from our list
    quit := make(chan int)        // Notify goroutines to quit

    // Start tracker goroutines
    for i := range to.Trackers {
        go to.Trackers[i].Run(peers, quit)
    }

    // Populate work queue
    for i := 0; i < to.Info.TotalPieces; i++ {
        work <- i
    }

    for {
        if to.shutdown {
            close(quit)
            return
        }
        select {
        case newPeer := <-peers:
            to.Peers = append(to.Peers, newPeer)
            go newPeer.StartWork(work, quit)
        case deadPeer := <-remove:
            // TODO close deadPeer
            to.removePeer(deadPeer)
        }
    }
}
