package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestTrackerPrint(t *testing.T) {
    assert := assert.New(t)

    var ta Tracker
    fmt.Println(ta)
    assert.NotNil(ta)
}
