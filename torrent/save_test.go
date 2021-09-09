package torrent

import (
	"context"
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
