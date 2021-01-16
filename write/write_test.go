package write

import (
    "testing"
    "fmt"

    "github.com/kylec725/graytorrent/torrent"
    "github.com/stretchr/testify/assert"
)

const debugWrite = false

func TestNewWrite(t *testing.T) {
    assert := assert.New(t)

    to := torrent.Torrent{Path: "../tmp/change.torrent"}

    err := NewWrite(to)
    assert.Nil(err)
    if debugWrite {
        fmt.Println("NewWrite error:", err)
    }
}
