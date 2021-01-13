package torrent

import (
    "testing"
    "fmt"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/stretchr/testify/assert"
)

func TestGetTrackers(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Name: "../tmp/change.torrent"}
    meta, err := metainfo.GetMeta(to.Name)
    if assert.Nil(err) {
        assert.NotNil(meta)
    }

    to.Trackers, err = getTrackers(meta)
    if assert.Nil(err) {
        for _, tr := range to.Trackers {
            assert.NotNil(tr)
            // fmt.Println(tr)
        }
    }
}

func TestBuildURL(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Name: "../tmp/1056.txt.utf-8.torrent"}
    to.Setup()
    meta, _ := metainfo.GetMeta(to.Name)

    for _, tr := range to.Trackers {
        assert.NotNil(tr)
        // Build url with each tracker object
        url, err := tr.buildURL(to.InfoHash, to.PeerID, 6881, meta.Info.Length, "started")
        if err == nil {
            // fmt.Println("tracker request:", url)
            assert.Equal("http://poole.cs.umd.edu:6969/announce?compact=1&downloaded=0&event=started&info_hash=Q%CB%DD%21%F2FYx%DAc%F0%91%B1y%18g2%CCX%05&left=810121&peer_id=" + string(to.PeerID[:]) + "&port=6881&uploaded=0", url, "Built tracker request was incorrect")
        }
    }
}

func TestGetPeers(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Name: "../tmp/batonroad.torrent"}
    to.Setup()
    meta, _ := metainfo.GetMeta(to.Name)

    var testTracker Tracker
    var testURL string
    for _, tr := range to.Trackers {
        assert.NotNil(tr)

        // Build url with each tracker object
        url, err := tr.buildURL(to.InfoHash, to.PeerID, 6881, meta.GetLength(), "stopped")
        assert.Nil(err)
        
        if tr.Announce[0:4] == "http" {
            testTracker = tr
            testURL = url
        }
    }
    // fmt.Println(testTracker)
    fmt.Println(testURL)

    _, err := testTracker.getPeers(testURL)
    assert.Nil(err)
}
