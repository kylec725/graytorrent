/*
Package metainfo provides functionality to extract metainfo about
torrents from their respective torrent files.
*/
package metainfo

import (
    "os"
    "strconv"
    "bytes"
    "crypto/sha1"

    "github.com/pkg/errors"
    bencode "github.com/jackpal/bencode-go"
)

// Errors
var (
    ErrPieceHashes = errors.New("Got malformed pieces from metainfo")
)

// BencodeMeta stores metainfo about a torrent file
type BencodeMeta struct {
    Info bencodeInfo `bencode:"info"`
    Announce string `bencode:"announce"`
    AnnounceList [][]string `bencode:"announce-list"`
}

type bencodeInfo struct {
    Name string `bencode:"name"`
    PieceLength int `bencode:"piece length"`
    Pieces string `bencode:"pieces"`
    Length int `bencode:"length,omitempty"` // Single file mode
    Files []bencodeFile `bencode:"files,omitempty"` // Multiple file mode
    Private int `bencode:"private,omitempty"` // Only use peers from tracker
}

type bencodeFile struct {
    Length int `bencode:"length"`
    Path []string `bencode:"path"`
}

func (meta BencodeMeta) String() string {
    var result string
    result += "Name: " + meta.Info.Name + "\n"
    result += "Announce: " + meta.Announce + "\n"
    for _, group := range meta.AnnounceList {
        for _, addr := range group {
            result += "Announce: " + addr + "\n"
        }
    }
    result += "PieceLength: " + strconv.Itoa(meta.Info.PieceLength) + "\n"

    totalLen, paths := meta.Info.Length, ""
    for _, file := range meta.Info.Files {
        totalLen += file.Length
        paths += file.Path[0] + " "
    }
    result += "Length: " + strconv.Itoa(totalLen) + "\n"
    if paths != "" {
        result += "Paths: " + paths + "\n"
    }

    return result
}

// Meta grabs bencoded metainfo and stores it into the BencodeMeta struct
func Meta(filename string) (BencodeMeta, error) {
    file, err := os.Open(filename)
    if err != nil {
        return BencodeMeta{}, errors.Wrap(err, "Meta")
    }
    defer file.Close()

    var meta BencodeMeta
    err = bencode.Unmarshal(file, &meta)
    if err != nil {
        return BencodeMeta{}, errors.Wrap(err, "Meta")
    }

    return meta, nil
}

// Length returns the total torrent length
func (meta BencodeMeta) Length() int {
    totalLen := meta.Info.Length
    for _, file := range meta.Info.Files {
        totalLen += file.Length
    }
    return totalLen
}

// InfoHash generates the infohash of the torrent file
func (meta BencodeMeta) InfoHash() ([20]byte, error) {
    var serialInfo bytes.Buffer
    err := bencode.Marshal(&serialInfo, meta.Info)
    if err != nil {
        return [20]byte{}, errors.Wrap(err, "InfoHash")
    }
    infoHash := sha1.Sum(serialInfo.Bytes())

    return infoHash, nil
}

// PieceHashes returns an array of the piece hashes of the torrent
func (meta BencodeMeta) PieceHashes() ([][20]byte, error) {
    piecesBytes := []byte(meta.Info.Pieces)
    if len(piecesBytes) % 20 != 0 {
        return nil, errors.Wrap(ErrPieceHashes, "Piecehashes")
    }

    totalPieces := len(piecesBytes) / 20
    pieceHashes := make([][20]byte, totalPieces)

    for i := 0; i < totalPieces; i++ {
        copy(pieceHashes[i][:], piecesBytes[ 20*i : 20*(i+1) ])
    }

    return pieceHashes, nil
}
