package torrent

import (
    "testing"
    "fmt"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/stretchr/testify/assert"
)

const debugTracker = false

func TestGetTrackers(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Source: "../tmp/change.torrent"}
    meta, err := metainfo.Meta(to.Source)
    assert.Nil(err, "Error with metainfo.Meta()")

    to.Trackers, err = getTrackers(meta)
    if assert.Nil(err) {
        for _, tr := range to.Trackers {
            assert.NotNil(tr)
            if debugTracker {
                fmt.Println(tr)
            }
        }
    }
}

func TestBuildURL(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Source: "../tmp/1056.txt.utf-8.torrent"}
    to.Setup()
    meta, _ := metainfo.Meta(to.Source)

    for _, tr := range to.Trackers {
        assert.NotNil(tr)
        // Build url with each tracker object
        url, err := tr.buildURL(to.InfoHash, to.PeerID, 6881, meta.Info.Length, "started")
        if err == nil {
            if debugTracker {
                fmt.Println("tracker request:", url)
            }
            assert.Equal("http://poole.cs.umd.edu:6969/announce?compact=1&downloaded=0&event=started&info_hash=Q%CB%DD%21%F2FYx%DAc%F0%91%B1y%18g2%CCX%05&left=810121&peer_id=" + string(to.PeerID[:]) + "&port=6881&uploaded=0", url, "Built tracker request was incorrect")
        }
    }
}
