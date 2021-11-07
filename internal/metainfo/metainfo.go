/*
Package metainfo provides functionality to extract metainfo about
torrents from their respective torrent files.
*/
package metainfo

import (
	"bytes"
	"crypto/sha1"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
	"github.com/pkg/errors"
)

// Errors
var (
	ErrPieceHashes = errors.New("Got malformed pieces from metainfo")
)

// Metainfo stores metainfo about a torrent file
type Metainfo struct {
	Info         bencodeInfo `bencode:"info"`
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"`
}

type bencodeInfo struct {
	Name        string        `bencode:"name"`
	PieceLength int           `bencode:"piece length"`
	Pieces      string        `bencode:"pieces"`
	Length      int           `bencode:"length,omitempty"`  // Single file mode
	Files       []bencodeFile `bencode:"files,omitempty"`   // Multiple file mode
	Private     int           `bencode:"private,omitempty"` // Only use peers from tracker
}

type bencodeFile struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

func (m Metainfo) String() string {
	var result string
	result += "Name: " + m.Info.Name + "\n"
	result += "Announce: " + m.Announce + "\n"
	for _, group := range m.AnnounceList {
		for _, addr := range group {
			result += "Announce: " + addr + "\n"
		}
	}
	result += "PieceLength: " + strconv.Itoa(m.Info.PieceLength) + "\n"

	totalLen, paths := m.Info.Length, ""
	for _, file := range m.Info.Files {
		totalLen += file.Length
		paths += file.Path[0] + " "
	}
	result += "Length: " + strconv.Itoa(totalLen) + "\n"
	if paths != "" {
		result += "Paths: " + paths + "\n"
	}

	return result
}

// New grabs bencoded metainfo and stores it into the Metainfo struct
func New(filename string) (Metainfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Metainfo{}, errors.Wrap(err, "Meta")
	}
	defer file.Close()

	var m Metainfo
	err = bencode.Unmarshal(file, &m)
	if err != nil {
		return Metainfo{}, errors.Wrap(err, "Meta")
	}

	return m, nil
}

// Length returns the total torrent length
func (m Metainfo) Length() int {
	totalLen := m.Info.Length
	for _, file := range m.Info.Files {
		totalLen += file.Length
	}
	return totalLen
}

// InfoHash generates the infohash of the torrent file
func (m Metainfo) InfoHash() ([20]byte, error) {
	var serialInfo bytes.Buffer
	err := bencode.Marshal(&serialInfo, m.Info)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "InfoHash")
	}
	infoHash := sha1.Sum(serialInfo.Bytes())

	return infoHash, nil
}

// PieceHashes returns an array of the piece hashes of the torrent
func (m Metainfo) PieceHashes() ([][20]byte, error) {
	piecesBytes := []byte(m.Info.Pieces)
	if len(piecesBytes)%20 != 0 {
		return nil, errors.Wrap(ErrPieceHashes, "Piecehashes")
	}

	totalPieces := len(piecesBytes) / 20
	pieceHashes := make([][20]byte, totalPieces)

	for i := 0; i < totalPieces; i++ {
		copy(pieceHashes[i][:], piecesBytes[20*i:20*(i+1)])
	}

	return pieceHashes, nil
}
