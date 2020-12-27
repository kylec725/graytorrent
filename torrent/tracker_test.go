package torrent

import (
    "testing"
    "fmt"

    // "github.com/stretchr/testify/assert"
)

func TestTrackerPrint(t *testing.T) {
    var testID [20]byte
    for i, c := range "--GT0100--abcdefghij" {
        testID[i] = byte(c)
    }
    tr := Tracker{"example.com/announce", false, 120, testID}
    fmt.Println(tr)
}
