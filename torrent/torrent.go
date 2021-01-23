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
    // TODO figure out how to remove a peer from the list if it has disconnected
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
    to.Trackers, err = getTrackers(meta)
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

// Send request to trackers concurrently to get list of peers
// func (to *Torrent) createStarted() {
//     for _, tr := range to.Trackers {
//         bytesLeft := (to.TotalPieces - to.Progress) * to.PieceLength
//         url, err := tr.buildURL(to.InfoHash, to.PeerID, 6881, bytesLeft, "started")
//         if err != nil {
//             log.Println("Error creating URL for:", tr.Announce)
//             continue
//         }
//     }
// }

// Download starts a routine to download a torrent from peers
// TODO
func (to *Torrent) Download() {
    // Start tracker goroutines

    peers := make(chan peer.Peer)
    work := make(chan int)
    remove := make(chan string)  // Channel for peer to notify it should be removed

    // Populate work queue
    for i := 0; i < to.Info.TotalPieces; i++ {
        work <- i
    }

    for {
        select {
        case newPeer := <-peers:
            to.Peers = append(to.Peers, newPeer)
        case deadPeer := <-remove:
            // TODO close deadPeer
            to.removePeer(deadPeer)
        }
    }
}

// connect to all peers asynchronously
// aynschronously add peers to an active peer list
// then use this peer list start requesting/getting pieces
