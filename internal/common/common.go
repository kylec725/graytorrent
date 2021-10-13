/*
Package common contains common structs required by several packages
in this project.
*/
package common

import (
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kylec725/graytorrent/internal/bitfield"
	"github.com/kylec725/graytorrent/internal/metainfo"
)

const peerID = "-GR0100-"

var (
	// GrayTorrentPath is the config directory of graytorrent
	GrayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")
	// SavePath is the directory to store data about managed torrents
	SavePath = filepath.Join(GrayTorrentPath, ".torrents")
)

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
	sync.Mutex
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

// SetPeerID sets a new peer ID
func (info *TorrentInfo) SetPeerID() {
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

// GetPaths retrieves the paths from metainfo
func GetPaths(meta metainfo.BencodeMeta) []Path {
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
func (info *TorrentInfo) PieceSize(index int) int {
	if index == info.TotalPieces-1 {
		return info.TotalLength - (info.TotalPieces-1)*info.PieceLength
	}
	return info.PieceLength
}
