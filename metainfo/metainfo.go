package metainfo

import (
    "log"
    "os"
    "strconv"
    "bytes"
    "crypto/sha1"
    // "fmt"

    bencode "github.com/jackpal/bencode-go"
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
    Length int `bencode:"length,omitempty"` // Single File Mode
    Files []bencodeFile `bencode:"files,omitempty"` // Multiple File Mode
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
    for _, addr := range meta.AnnounceList {
        result += "Announce: " + addr[0] + "\n"
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

// GetMeta grabs bencoded metainfo and stores it into the BencodeMeta struct
func GetMeta(filename string) (BencodeMeta, error) {
    file, err := os.Open(filename)
    if err != nil {
        log.Println("Could not open file:", filename)
        return BencodeMeta{}, err
    }
    defer file.Close()

    var meta BencodeMeta
    err = bencode.Unmarshal(file, &meta)
    if err != nil {
        log.Println("Could not unmarshal bencoded file:", filename)
        return BencodeMeta{}, err
    }

    return meta, nil
}

// GetInfoHash generates the infohash of the torrent file
func GetInfoHash(meta BencodeMeta) ([20]byte, error) {
    var serialInfo bytes.Buffer
    err := bencode.Marshal(&serialInfo, meta.Info)
    if err != nil {
        return [20]byte{}, err
    }
    infoHash := sha1.Sum(serialInfo.Bytes())

    return infoHash, nil
}
