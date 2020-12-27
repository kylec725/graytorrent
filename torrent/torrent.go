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
    // "log"
    "time"
    "math/rand"
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
    Filename string
    Trackers []Tracker
    PieceLength int
    InfoHash [20]byte
    PieceHashes [][20]byte
    ID [20]byte
}

func (to *Torrent) setID() {
    rand.Seed(time.Now().UnixNano())
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    id := "-GT0100-"

    for i := 0; i < 12; i++ {
        pos := rand.Intn(len(chars))
        id += string(chars[pos])
    }

    for i, c := range id {
        to.ID[i] = byte(c)
    }
}

func (to *Torrent) setInfoHash() {

}
