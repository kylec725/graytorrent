package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestTrackerPrint(t *testing.T) {
    var testID [20]byte
    for i, c := range "--GT0100--abcdefghij" {
        testID[i] = byte(c)
    }
    tr := Tracker{"example.com/announce", false, 120, testID}
    fmt.Println(tr)
}

func TestGetTrackers(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Filename: "../tmp/change.torrent"}
    meta, err := to.getMeta()
    if assert.Nil(err) {
        assert.NotNil(meta)
    }

    trackers, err := to.getTrackers()
    if assert.Nil(err) {
        for _, tr := range trackers {
            assert.NotNil(tr)
            fmt.Println(tr)
        }
    }
}
