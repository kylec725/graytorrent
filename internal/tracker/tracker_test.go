package tracker

import (
	"fmt"
	"testing"

	"github.com/kylec725/gray/internal/metainfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const debugTracker = false

func TestGetTrackers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	meta, err := metainfo.Meta("../tmp/change.torrent")
	require.Nil(err, "Error with metainfo.Meta()")
	// info, err := common.GetInfo(meta)
	// require.Nil(err, "GetInfo() error")

	trackers, err := GetTrackers(meta)

	if assert.Nil(err) {
		for _, tr := range trackers {
			assert.NotNil(tr)
			if debugTracker {
				fmt.Println(tr)
			}
		}
	}
}
