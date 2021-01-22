package torrent

import (
    "testing"
    "fmt"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/stretchr/testify/assert"
)

const debugRequests = false

func TestTrackerReqs(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Path: "../tmp/batonroad.torrent"}
    to.Setup()
    meta, _ := metainfo.Meta(to.Path)

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

    peersList, err := testTracker.sendStarted(&to.Info, 6881, meta.Length())
    // Retry a get request since this tracker seems to fail on the first connection attempt
    if err != nil {
        peersList, err = testTracker.sendStarted(&to.Info, 6881, meta.Length())
    }
    if assert.Nil(err) {
        for _, peer := range peersList {
            if debugRequests {
                fmt.Println("Peer:", peer)
            }
        }
        err = testTracker.sendStopped(&to.Info, 6881, meta.Length())
        assert.Nil(err)

        if debugRequests {
            fmt.Printf("Tracker%+v\n", testTracker)
        }
    }

}
