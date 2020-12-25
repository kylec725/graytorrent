package torrent

import (
    "testing"
    "fmt"
)

func TestMetaSingle(t *testing.T) {
    var to Torrent = Torrent{Filename: "../tmp/1056.txt.utf-8.torrent"}
    meta, err := to.read()
    if err != nil {
        t.Error("Error getting single metainfo:", err)
    }
    fmt.Println(meta)
}

func TestMetaMulti(t *testing.T) {
    var to Torrent = Torrent{Filename: "../tmp/shared.torrent"}
    meta, err := to.read()
    if err != nil {
        t.Error("Error getting multi metainfo:", err)
    }
    fmt.Println(meta)
}
