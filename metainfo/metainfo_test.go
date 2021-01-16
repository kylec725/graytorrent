package metainfo

import (
    "testing"
    "encoding/hex"
    "fmt"

    "github.com/stretchr/testify/assert"
)

const debugMetainfo = false

func TestMetaBasic(t *testing.T) {
    assert := assert.New(t)

    meta, err := Meta("../tmp/1056.txt.utf-8.torrent")
    if assert.Nil(err) {
        if debugMetainfo {
            fmt.Println(meta)
        }
        assert.NotNil(meta)
    }
}

func TestMetaExtra(t *testing.T) {
    assert := assert.New(t)

    meta, err := Meta("../tmp/shared.torrent")
    if assert.Nil(err) {
        if debugMetainfo {
            fmt.Println(meta)
        }
        assert.NotNil(meta)
    }

    meta, err = Meta("../tmp/change.torrent")
    if assert.Nil(err) {
        if debugMetainfo {
            fmt.Println(meta)
        }
        assert.NotNil(meta)
    }
}

func TestInfoHash(t *testing.T) {
    assert := assert.New(t)

    meta, err := Meta("../tmp/1056.txt.utf-8.torrent")
    if assert.Nil(err) {
        infoHash, err := meta.InfoHash()
        if assert.Nil(err) {
            if debugMetainfo {
                fmt.Println("infohash:", hex.EncodeToString(infoHash[:]))
            }
            assert.Equal("51cbdd21f2465978da63f091b179186732cc5805", hex.EncodeToString(infoHash[:]), "Calculated the info hash incorrectly")
        }
    }

    meta, err = Meta("../tmp/change.torrent")
    if assert.Nil(err) {
        infoHash, err := meta.InfoHash()
        if assert.Nil(err) {
            if debugMetainfo {
                fmt.Println("infohash:", hex.EncodeToString(infoHash[:]))
            }
            assert.Equal("74df948ea813e7938a207b0bb23d0edf2b74f4b1", hex.EncodeToString(infoHash[:]), "Calculated the info hash incorrectly")
        }
    }

    meta, err = Meta("../tmp/batonroad.torrent")
    if assert.Nil(err) {
        infoHash, err := meta.InfoHash()
        if assert.Nil(err) {
            if debugMetainfo {
                fmt.Println("infohash:", hex.EncodeToString(infoHash[:]))
            }
            assert.Equal("de22c582d9958b6b53d3cb1643c3f7dd4a0930f4", hex.EncodeToString(infoHash[:]), "Calculated the info hash incorrectly")
        }
    }
}

func TestPieceHashes(t *testing.T) {
    assert := assert.New(t)

    meta, err := Meta("../tmp/1056.txt.utf-8.torrent")
    if assert.Nil(err) {
        pieceHashes, err := meta.PieceHashes()
        if assert.Nil(err) {
            assert.NotNil(pieceHashes)
        }
    }
}
