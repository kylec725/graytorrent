package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

const debugTorrent = false

func TestGetID(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Name: "../tmp/change.torrent"}
    to.setID()
    halfID := string(to.PeerID[0:8])

    if debugTorrent {
        fmt.Println("ID:", to.PeerID)
    }

    assert.Equal("-GT0100-", halfID, "First half of ID was not set correctly")
}
