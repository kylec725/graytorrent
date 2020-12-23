package torrent_test

import (
	"testing"

	"github.com/kylec725/graytorrent/torrent"
)

func TestTorrentPrint(t *testing.T) {
	var to torrent.Torrent = 7
	to.Print()
}
