package torrent_test

import (
    "testing"

    "github.com/kylec725/graytorrent/torrent"
)

func TestPrint(t *testing.T) {
    torrent.TorrentPrint()
    torrent.TrackerPrint()
}
