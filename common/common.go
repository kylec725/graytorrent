/*
Package common contains common structs required by several packages
in this project.
*/
package common

import (
	"math"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/kylec725/graytorrent/bitfield"
	"github.com/kylec725/graytorrent/metainfo"
	"github.com/pkg/errors"
)

const peerID = "-GT0100-"

// TorrentInfo contains information about a torrent
type TorrentInfo struct {
	Name        string            `json:"Name"`
	Paths       []Path            `json:"Paths"`
	Bitfield    bitfield.Bitfield `json:"Bitfield"`    // bitfield of current pieces
	PieceLength int               `json:"PieceLength"` // number of bytes per piece
	TotalPieces int               `json:"TotalPieces"` // total pieces in the torrent
	TotalLength int               `json:"TotalLength"` // total length of the torrent
	Left        int               `json:"Left"`        // number of bytes left to torrent
	InfoHash    [20]byte          `json:"InfoHash"`
	PieceHashes [][20]byte        `json:"PieceHashes"`
	PeerID      [20]byte          `json:"PeerID"`
}

// Path stores info about each file in a torrent
type Path struct {
	Length int    `json:"Length"`
	Path   string `json:"Path"`
}

// Min returns the minimum of two integers
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
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
	info.Left = info.TotalLength

	// Initialize the bitfield
	bitfieldSize := int(math.Ceil(float64(info.TotalPieces) / 8))
	info.Bitfield = make([]byte, bitfieldSize)

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
		paths[0] = Path{Length: meta.Info.Length, Path: meta.Info.Name}
		return paths
	}

	// Multiple files
	var paths []Path
	for _, file := range meta.Info.Files {
		newPath := filepath.Join(file.Path...)
		newPath = filepath.Join(meta.Info.Name, newPath)
		paths = append(paths, Path{Length: file.Length, Path: newPath})
	}

	return paths
}

// PieceSize returns the size of a piece at a specified index
func PieceSize(info TorrentInfo, index int) int {
	if index == info.TotalPieces-1 {
		return info.TotalLength - (info.TotalPieces-1)*info.PieceLength
	}
	return info.PieceLength
}
