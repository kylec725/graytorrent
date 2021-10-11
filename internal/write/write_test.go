package write

import (
	"fmt"
	"os"
	"testing"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/metainfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const debugWrite = false

func TestNewWriteSingle(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Setup
	meta, err := metainfo.Meta("../tmp/change.torrent")
	require.Nil(err, "Meta() error")
	info, err := common.GetInfo(meta)
	require.Nil(err, "GetInfo() error")

	// Remove the torrent's filename if it exists
	if _, err := os.Stat(info.Name); err == nil {
		err = os.Remove(info.Name)
		if err != nil {
			panic("Removing test file failed")
		}
	}

	err = NewWrite(info)
	if assert.Nil(err) {
		if debugWrite {
			fmt.Println("File created:", info.Name)
		}

		// Test that creating an identical file throws an error
		err = NewWrite(info)
		assert.NotNil(err)
	}
}

func TestNewWriteMulti(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Setup
	meta, err := metainfo.Meta("../tmp/batonroad.torrent")
	require.Nil(err, "Meta() error")
	info, err := common.GetInfo(meta)
	require.Nil(err, "GetInfo() error")

	// Remove the torrent's filename if it exists
	if _, err := os.Stat(info.Name); err == nil {
		err = os.RemoveAll(info.Name)
		if err != nil {
			panic("Removing test file failed")
		}
	}

	err = NewWrite(info)
	if assert.Nil(err) {
		if debugWrite {
			fmt.Println("File created:", info.Name)
		}

		// Test that creating an identical file throws an error
		err = NewWrite(info)
		assert.NotNil(err)
	}
}

func TestAddBlock(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Setup
	meta, err := metainfo.Meta("../tmp/change.torrent")
	require.Nil(err, "Meta() error")
	info, err := common.GetInfo(meta)
	require.Nil(err, "GetInfo() error")

	index := 8
	begin := 0
	piece := make([]byte, info.PieceLength)
	block := []byte("hello")

	if debugWrite {
		fmt.Println("BlockLength:", len(block))
		fmt.Println("PieceLength:", len(piece))
	}

	err = AddBlock(info, index, begin, block, piece)
	if assert.Nil(err) {
		assert.Equal(block, piece[begin:begin+len(block)])
	}
}

func TestAddPiece(t *testing.T) {
	assert := assert.New(t)

	info := &common.TorrentInfo{
		Name:        "test",
		PieceLength: 5,
		TotalLength: 19,
		TotalPieces: 4,
		Paths: []common.Path{
			{Length: 2, Path: "test/0.txt"},
			{Length: 2, Path: "test/1.txt"},
			{Length: 1, Path: "test/2.txt"},
			{Length: 5, Path: "test/3.txt"},
			{Length: 9, Path: "test/4.txt"},
		},
	}
	// Remove the torrent's filename if it exists
	if _, err := os.Stat(info.Name); err == nil {
		err = os.RemoveAll(info.Name)
		if err != nil {
			panic("Removing test file failed")
		}
	}
	err := NewWrite(info)
	assert.Nil(err, "NewWrite error")

	index := 0
	piece := []byte("00112")
	err = AddPiece(info, index, piece)
	assert.Nil(err)
	if debugWrite {
		fmt.Printf("wrote piece %d: %s\n", index, string(piece))
	}

	index = 1
	piece = []byte("33333")
	err = AddPiece(info, index, piece)
	assert.Nil(err)
	if debugWrite {
		fmt.Printf("wrote piece %d: %s\n", index, string(piece))
	}

	index = 2
	piece = []byte("44444")
	err = AddPiece(info, index, piece)
	assert.Nil(err)
	if debugWrite {
		fmt.Printf("wrote piece %d: %s\n", index, string(piece))
	}

	index = 3
	piece = []byte("4444")
	err = AddPiece(info, index, piece)
	assert.Nil(err)
	if debugWrite {
		fmt.Printf("wrote piece %d: %s\n", index, string(piece))
	}
}

// Needs TestAddPiece to work first
func TestReadPiece(t *testing.T) {
	assert := assert.New(t)

	info := &common.TorrentInfo{
		Name:        "test",
		PieceLength: 5,
		TotalLength: 19,
		TotalPieces: 4,
		Paths: []common.Path{
			{Length: 2, Path: "test/0.txt"},
			{Length: 2, Path: "test/1.txt"},
			{Length: 1, Path: "test/2.txt"},
			{Length: 5, Path: "test/3.txt"},
			{Length: 9, Path: "test/4.txt"},
		},
	}

	for index := 0; index < info.TotalPieces; index++ {
		piece, err := ReadPiece(info, index)
		assert.Nil(err)
		if debugWrite {
			fmt.Printf("read piece %d: %s\n", index, string(piece))
		}
	}
}
