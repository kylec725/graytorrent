package tracker

import (
	"fmt"
	"testing"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/metainfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const debugUDP = false

func TestUDPReqs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	meta, err := metainfo.Meta("../tmp/batonroad.torrent")
	require.Nil(err, "Error with metainfo.Meta()")
	info, err := common.GetInfo(meta)
	require.Nil(err, "GetInfo() error")

	trackers, err := GetTrackers(meta)

	var testTracker Tracker
	for _, tr := range trackers {
		assert.NotNil(tr)

		if tr.Announce[0:3] == "udp" {
			testTracker = tr
			break
		}
	}

	if debugUDP {
		fmt.Printf("UDP Tracker%+v\n", testTracker)
	}

	peersList, err := testTracker.sendStarted(info, 6881, 0, 0, info.Left)
	if assert.Nil(err) {
		for _, peer := range peersList {
			if debugUDP {
				fmt.Println("UDP Peer:", peer)
			}
		}
		err = testTracker.sendStopped(info, 6881, 0, 0, info.Left)
		assert.Nil(err)

		if debugUDP {
			fmt.Printf("UDP Tracker%+v\n", testTracker)
		}
	}

}
