package torrent

import (
    "testing"
    "fmt"
)

func TestMeta(t *testing.T) {
    var to Torrent = Torrent{Filename: "../tmp/1056.txt.utf-8.torrent"}
    meta, err := to.getMeta()
    if err != nil {
        t.Error("Error getting single metainfo:", err)
    }
    fmt.Println(meta)

    to = Torrent{Filename: "../tmp/shared.torrent"}
    meta, err = to.getMeta()
    if err != nil {
        t.Error("Error getting multi metainfo:", err)
    }
    fmt.Println(meta)

    to = Torrent{Filename: "../tmp/change.torrent"}
    meta, err = to.getMeta()
    if err != nil {
        t.Error("Error getting multiple announce metainfo:", err)
    }
    fmt.Println(meta)
}
