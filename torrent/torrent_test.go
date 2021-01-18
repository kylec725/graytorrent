package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

const debugTorrent = false

func TestSetup(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Source: "../tmp/change.torrent"}
    err := to.Setup()
    if assert.Nil(err) {
        assert.Equal("[Nipponsei] BLEACH OP12 Single - chAngE [miwa].zip", to.Name, "Name is incorrect")
        assert.Equal(262144, to.PieceLength, "Name is incorrect")
        assert.Equal(150, to.TotalPieces, "Name is incorrect")
    }
}

func TestGetID(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Source: "../tmp/change.torrent"}
    to.setID()
    halfID := string(to.PeerID[0:8])

    if debugTorrent {
        fmt.Println("ID:", to.PeerID)
    }

    assert.Equal("-GT0100-", halfID, "First half of ID was not set correctly")
}
