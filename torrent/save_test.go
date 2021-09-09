package torrent

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kylec725/graytorrent/common"
	"github.com/stretchr/testify/assert"
)

const debugSave = false

func TestSave(t *testing.T) {
	assert := assert.New(t)

	ctx := context.WithValue(context.Background(), common.KeyPort, uint16(6881))
	var to Torrent = Torrent{Path: "../tmp/change.torrent"}
	err := to.Setup(ctx)
	if assert.Nil(err) {
		err = to.Save()
		assert.Nil(err)
		os.Remove(to.Info.Name)
	}
}

func TestLoadAll(t *testing.T) {
	assert := assert.New(t)

	torrentList, err := LoadAll()
	if assert.Nil(err) {
		assert.NotEmpty(torrentList)
		if debugSave {
			for _, to := range torrentList {
				fmt.Println("Name:", to.Info.Name)
				fmt.Println("PieceLength:", to.Info.PieceLength)
				fmt.Println("TotalPieces:", to.Info.TotalPieces)
				fmt.Println("TotalLength:", to.Info.TotalLength)
				fmt.Println("Left:", to.Info.Left)
			}
		}
	}
}
