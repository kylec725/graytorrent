package torrent

import (
    "testing"
    "fmt"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/metainfo"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

const debugTracker = false

func TestGetTrackers(t *testing.T) {
    assert := assert.New(t)
    require := require.New(t)

    to := Torrent{Path: "../tmp/change.torrent"}
    meta, err := metainfo.Meta(to.Path)
    require.Nil(err, "Error with metainfo.Meta()")
    info, err := common.GetInfo(meta)
    require.Nil(err, "GetInfo() error")

    to.Trackers, err = getTrackers(meta, &info, 6881)
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

    to := Torrent{Path: "../tmp/1056.txt.utf-8.torrent"}
    to.Setup()
    meta, _ := metainfo.Meta(to.Path)

    for _, tr := range to.Trackers {
        assert.NotNil(tr)
        // Build url with each tracker object
        url, err := tr.buildURL(to.Info.InfoHash, to.Info.PeerID, 6881, meta.Info.Length, "started")
        if err == nil {
            if debugTracker {
                fmt.Println("tracker request:", url)
            }
            assert.Equal("http://poole.cs.umd.edu:6969/announce?compact=1&downloaded=0&event=started&info_hash=Q%CB%DD%21%F2FYx%DAc%F0%91%B1y%18g2%CCX%05&left=810121&peer_id=" + string(to.Info.PeerID[:]) + "&port=6881&uploaded=0", url, "Built tracker request was incorrect")
        }
    }
}
