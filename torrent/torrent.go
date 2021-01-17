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
    "time"
    "math/rand"
    "path/filepath"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/pkg/errors"
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
    Source string
    Name string
    Paths []Path
    Trackers []Tracker
    Progress int // total number of pieces we have
    PieceLength int // number of bytes per piece
    TotalPieces int // total pieces in the torrent
    InfoHash [20]byte
    PieceHashes [][20]byte
    PeerID [20]byte
}

// Path stores info about each file in a torrent
type Path struct {
    length int
    path string
}

// Setup gets and sets up necessary properties of a new torrent object
func (to *Torrent) Setup() error {
    // Get metainfo
    meta, err := metainfo.Meta(to.Source)
    if err != nil {
        return errors.Wrap(err, "Setup")
    }

    // Set torrent name
    to.Name = meta.Info.Name

    // Set info about file length
    to.Progress = 0
    to.PieceLength = meta.Info.PieceLength
    to.TotalPieces = len(meta.Info.Pieces) / 20

    // Set torrent's filepaths
    to.Paths = getPaths(meta)

    // Set the peer ID
    to.setID()

    // Get the infohash from the metainfo
    to.InfoHash, err = meta.InfoHash()
    if err != nil {
        return errors.Wrap(err, "Setup")
    }

    // Get the piece hashes from the metainfo
    to.PieceHashes, err = meta.PieceHashes()
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

func (to *Torrent) setID() {
    rand.Seed(time.Now().UnixNano())
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    id := "-GT0100-"

    for i := 0; i < 12; i++ {
        pos := rand.Intn(len(chars))
        id += string(chars[pos])
    }

    for i, c := range id {
        to.PeerID[i] = byte(c)
    }
}

func getPaths(meta metainfo.BencodeMeta) []Path {
    // Single file
    if meta.Info.Length > 0 {
        paths := make([]Path, 1)
        paths[0] = Path{ length: meta.Info.Length, path: meta.Info.Name }
        return paths
    }

    // Multiple files
    var paths []Path
    for _, file := range meta.Info.Files {
        newPath := filepath.Join(file.Path...)
        paths = append(paths, Path{ length: file.Length, path: newPath })
    }

    return paths
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
