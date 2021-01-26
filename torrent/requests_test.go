package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

const debugRequests = false

func TestTrackerReqs(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Path: "../tmp/batonroad.torrent", Port: 6881}
    to.Setup()

    var testTracker Tracker
    for _, tr := range to.Trackers {
        assert.NotNil(tr)

        if tr.Announce[0:4] == "http" {
            testTracker = tr
        }
    }

    if debugRequests {
        fmt.Printf("Tracker%+v\n", testTracker)
    }

    peersList, err := testTracker.sendStarted()
    if assert.Nil(err) {
        for _, peer := range peersList {
            if debugRequests {
                fmt.Println("Peer:", peer)
            }
        }
        err = testTracker.sendStopped()
        assert.Nil(err)

        if debugRequests {
            fmt.Printf("Tracker%+v\n", testTracker)
        }
    }

}
