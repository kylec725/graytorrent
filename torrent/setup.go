package torrent

import (
	"math"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/metainfo"
	"github.com/kylec725/graytorrent/internal/tracker"
	"github.com/pkg/errors"
)

// InfoFromFile grabs torrent metainfo from a .torrent file
func InfoFromFile(filename string) (*common.TorrentInfo, []*tracker.Tracker, error) {
	var info common.TorrentInfo

	meta, err := metainfo.New(filename)
	if err != nil {
		return nil, nil, errors.Wrap(err, "InfoFromFile")
	}

	info.Name = meta.Info.Name

	info.PieceLength = meta.Info.PieceLength
	info.TotalPieces = len(meta.Info.Pieces) / 20
	info.TotalLength = meta.Length()
	info.Left = info.TotalLength

	bitfieldSize := int(math.Ceil(float64(info.TotalPieces) / 8))
	info.Bitfield = make([]byte, bitfieldSize)

	// Set torrent's filepaths
	info.Paths = common.GetPaths(meta)

	info.SetPeerID() // TODO: Set peerID once for the client, and make it persistent

	info.InfoHash, err = meta.InfoHash()
	if err != nil {
		return nil, nil, errors.Wrap(err, "InfoFromFile")
	}

	// Get the piece hashes from the metainfo
	info.PieceHashes, err = meta.PieceHashes()
	if err != nil {
		return nil, nil, errors.Wrap(err, "InfoFromFile")
	}

	// Trackers
	trackers, err := tracker.GetTrackers(meta)
	if err != nil {
		return nil, nil, errors.Wrap(err, "InfoFromFile")
	}

	return &info, trackers, nil
}
