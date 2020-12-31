package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestGetID(t *testing.T) {
    assert := assert.New(t)

    var to Torrent = Torrent{Name: "../tmp/change.torrent"}
    to.setID()
    halfID := string(to.ID[0:8])

    fmt.Println("Generated ID:", string(to.ID[:]))
    fmt.Println()
    assert.Equal("-GT0100-", halfID, "First half of ID was not set correctly")
}
