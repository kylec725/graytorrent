package tracker

import (
	"fmt"
	"testing"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/metainfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const debugRequests = false

func TestTrackerReqs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	meta, err := metainfo.Meta("../tmp/batonroad.torrent")
	require.Nil(err, "Error with metainfo.Meta()")
	info, err := common.GetInfo(meta)
	require.Nil(err, "GetInfo() error")

	trackers, err := GetTrackers(meta, 6881)

	var testTracker Tracker
	for _, tr := range trackers {
		assert.NotNil(tr)

		if tr.Announce[0:4] == "http" {
			testTracker = tr
		}
	}

	if debugRequests {
		fmt.Printf("Tracker%+v\n", testTracker)
	}

	peersList, err := testTracker.sendStarted(info, 6881)
	if assert.Nil(err) {
		for _, peer := range peersList {
			if debugRequests {
				fmt.Println("Peer:", peer)
			}
		}
		err = testTracker.sendStopped(info, 6881)
		assert.Nil(err)

		if debugRequests {
			fmt.Printf("Tracker%+v\n", testTracker)
		}
	}

}
