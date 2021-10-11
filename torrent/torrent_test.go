package torrent

import (
	"os"
	"testing"

	// "fmt"
	"context"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/stretchr/testify/assert"
)

const debugTorrent = false

func TestSetup(t *testing.T) {
	assert := assert.New(t)

	ctx := context.WithValue(context.Background(), common.KeyPort, uint16(6881))
	var to Torrent = Torrent{File: "../tmp/change.torrent"}
	err := to.Setup(ctx)
	if assert.Nil(err) {
		assert.Equal("[Nipponsei] BLEACH OP12 Single - chAngE [miwa].zip", to.Info.Name, "Name is incorrect")
		assert.Equal(262144, to.Info.PieceLength, "PieceLength is incorrect")
		assert.Equal(150, to.Info.TotalPieces, "TotalPieces is incorrect")
		os.Remove(to.Info.Name)
	}
}
