package torrent

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestGetID(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Name: "../tmp/change.torrent"}
    to.setID()
    halfID := string(to.PeerID[0:8])

    assert.Equal("-GT0100-", halfID, "First half of ID was not set correctly")
}
