package torrent

import (
    "testing"
    "fmt"
)

func TestTorrentPrint(t *testing.T) {
    var to Torrent = Torrent{Filename: "tester"}
    fmt.Println(to)
}

func TestFileRead(t *testing.T) {
    var to Torrent = Torrent{Filename: "../tmp/one.txt"}
    to.read()
}
