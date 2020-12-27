package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Filename: "../tmp/1056.txt.utf-8.torrent"}
    meta, err := to.getMeta()
    if assert.Nil(err) {
        assert.NotNil(meta)
    }
    fmt.Println(meta)

    to = Torrent{Filename: "../tmp/shared.torrent"}
    meta, err = to.getMeta()
    if assert.Nil(err) {
        assert.NotNil(meta)
    }
    fmt.Println(meta)

    to = Torrent{Filename: "../tmp/change.torrent"}
    meta, err = to.getMeta()
    if assert.Nil(err) {
        assert.NotNil(meta)
    }
    fmt.Println(meta)
}

func TestGetID(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Filename: "../tmp/1056.txt.utf-8.torrent"}
    to.setID()
    halfID := string(to.ID[0:10])

    fmt.Println("Generated ID:", string(to.ID[:]))
    fmt.Println()
    assert.Equal("--GT0100--", halfID, "First half of ID was not set correctly")
}
