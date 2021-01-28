package torrent

import (
	"testing"
	// "fmt"
	"context"

	"github.com/kylec725/graytorrent/common"
	"github.com/stretchr/testify/assert"
)

const debugTorrent = false

func TestSetup(t *testing.T) {
	assert := assert.New(t)

	ctx := context.WithValue(context.Background(), common.KeyPort, uint16(6881))
	var to Torrent = Torrent{Path: "../tmp/change.torrent"}
	err := to.Setup(ctx)
	if assert.Nil(err) {
		assert.Equal("[Nipponsei] BLEACH OP12 Single - chAngE [miwa].zip", to.Info.Name, "Name is incorrect")
		assert.Equal(262144, to.Info.PieceLength, "Name is incorrect")
		assert.Equal(150, to.Info.TotalPieces, "Name is incorrect")
	}
}
