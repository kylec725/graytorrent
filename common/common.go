/*
Package common contains common structs required by several packages
in this project.
*/
package common

import (
    "math/rand"
    "time"
    "path/filepath"

    "github.com/kylec725/graytorrent/bitfield"
    "github.com/kylec725/graytorrent/metainfo"

    "github.com/pkg/errors"
)

const peerID = "-GT0100-"

// TorrentInfo contains information about a torrent
type TorrentInfo struct {
    Name string
    Paths []Path
    Bitfield bitfield.Bitfield  // bitfield of current pieces
    PieceLength int  // number of bytes per piece
    TotalPieces int  // total pieces in the torrent
    TotalLength int  // total length of the torrent
    InfoHash [20]byte
    PieceHashes [][20]byte
    PeerID [20]byte
}

// Path stores info about each file in a torrent
type Path struct {
    Length int
    Path string
}

// GetInfo uses metainfo to retrieve information about a torrent
func GetInfo(meta metainfo.BencodeMeta) (TorrentInfo, error) {
    var info TorrentInfo

    // Set torrent name
    info.Name = meta.Info.Name

    // Set info about file length
    info.PieceLength = meta.Info.PieceLength
    info.TotalPieces = len(meta.Info.Pieces) / 20
    info.TotalLength = meta.Length()

    // Set torrent's filepaths
    info.Paths = getPaths(meta)

    // Set the peer ID
    info.setID()

    // Get the infohash from the metainfo
    var err error
    info.InfoHash, err = meta.InfoHash()
    if err != nil {
        return TorrentInfo{}, errors.Wrap(err, "SetInfo")
    }

    // Get the piece hashes from the metainfo
    info.PieceHashes, err = meta.PieceHashes()
    if err != nil {
        return TorrentInfo{}, errors.Wrap(err, "SetInfo")
    }

    return info, nil
}

func (info *TorrentInfo) setID() {
    rand.Seed(time.Now().UnixNano())
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    id := peerID

    for i := 0; i < 12; i++ {
        pos := rand.Intn(len(chars))
        id += string(chars[pos])
    }

    for i, c := range id {
        info.PeerID[i] = byte(c)
    }
}

func getPaths(meta metainfo.BencodeMeta) []Path {
    // Single file
    if meta.Info.Length > 0 {
        paths := make([]Path, 1)
        paths[0] = Path{ Length: meta.Info.Length, Path: meta.Info.Name }
        return paths
    }

    // Multiple files
    var paths []Path
    for _, file := range meta.Info.Files {
        newPath := filepath.Join(file.Path...)
        newPath = filepath.Join(meta.Info.Name, newPath)
        paths = append(paths, Path{ Length: file.Length, Path: newPath })
    }

    return paths
}
