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
