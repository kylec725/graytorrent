package torrent

import (
    "testing"
    // "fmt"

    "github.com/stretchr/testify/assert"
)

func TestGetTrackers(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Filename: "../tmp/change.torrent"}
    meta, err := to.getMeta()
    if assert.Nil(err) {
        assert.NotNil(meta)
    }

    trackers, err := to.getTrackers()
    if assert.Nil(err) {
        for _, tr := range trackers {
            assert.NotNil(tr)
            // fmt.Println(tr)
        }
    }
}

func TestBuildURL(t *testing.T) {
    assert := assert.New(t)

    to := Torrent{Filename: "../tmp/1056.txt.utf-8.torrent"}
    meta, err := to.getMeta()
    if assert.Nil(err) {
        assert.NotNil(meta)
    }

    trackers, err := to.getTrackers()
    if assert.Nil(err) {
        for _, tr := range trackers {
            assert.NotNil(tr)
            // Build url with each tracker object
            // url := tr.buildURL()
            // assert.Equal("http://poole.cs.umd.edu:6969/announce?info_hash=Q%CB%DD%21%F2FYx%DAc%F0%91%B1y%18g2%CCX%05&peer_id=-GC1111-0IP4NCVJSIu3&port=6881&numwant=50&uploaded=0&downloaded=0&left=810121&compact=1&event=", url, "Built tracker request was incorrect")
        }
    }

}
