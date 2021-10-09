package torrent

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/stretchr/testify/assert"
)

const debugSave = false

func TestSaveAll(t *testing.T) {
	assert := assert.New(t)

	ctx := context.WithValue(context.Background(), common.KeyPort, uint16(6881))
	torrentList := make([]Torrent, 0)
	torrentList = append(torrentList, Torrent{Path: "../tmp/change.torrent"})
	torrentList = append(torrentList, Torrent{Path: "../tmp/1056.txt.utf-8.torrent"})
	torrentList = append(torrentList, Torrent{Path: "../tmp/1184-0.txt.torrent"})

	for i := range torrentList {
		err := torrentList[i].Setup(ctx)
		assert.Nil(err)
		os.Remove(torrentList[i].Info.Name)
	}

	SaveAll(torrentList)
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
