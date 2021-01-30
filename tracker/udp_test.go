package tracker

import (
	"fmt"
	"testing"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/metainfo"
	"github.com/stretchr/testify/require"
)

const debugUDP = true

func TestUDPReqs(t *testing.T) {
	// assert := assert.New(t)
	require := require.New(t)

	meta, err := metainfo.Meta("../tmp/batonroad.torrent")
	require.Nil(err, "Error with metainfo.Meta()")
	_, err = common.GetInfo(meta)
	require.Nil(err, "GetInfo() error")

	trackers, err := GetTrackers(meta, 6881)

	var testTracker Tracker
	testTracker = trackers[0]
	// for _, tr := range trackers {
	// 	assert.NotNil(tr)
	//
	// 	if tr.Announce[0:3] == "udp" {
	// 		testTracker = tr
	// 		break
	// 	}
	// }

	if debugUDP {
		fmt.Printf("Tracker%+v\n", testTracker)
	}

	// peersList, err := testTracker.sendStarted(info, 6881, 0, 0, info.Left)
	// if assert.Nil(err) {
	// 	for _, peer := range peersList {
	// 		if debugRequests {
	// 			fmt.Println("Peer:", peer)
	// 		}
	// 	}
	// 	err = testTracker.sendStopped(info, 6881, 0, 0, info.Left)
	// 	assert.Nil(err)
	//
	// 	if debugRequests {
	// 		fmt.Printf("Tracker%+v\n", testTracker)
	// 	}
	// }

}
