/*
Package torrent provides a library for reading from a torrent file
and storing desired information for leeching or seeding.
Tracker file to handle grabbing information about current
peers and the state of the file.
Write file to handle writing and getting pieces, as well as verifying
the hash of received pieces.
Will communicate with the peers package for sending and receiving
pieces of the torrent.
*/
package torrent

import (
    "log"
    "os"
    "strconv"
    "time"
    "math/rand"

    bencode "github.com/jackpal/bencode-go"
)

// Torrent stores metainfo and current progress on a torrent
type Torrent struct {
    Filename string
    Trackers []Tracker
    PieceLength int
    InfoHash [20]byte
    PieceHashes [][20]byte
    ID [20]byte
}

// Use struct with nested struct to decode the bencoded file
type bencodeMeta struct {
    Info bencodeInfo `bencode:"info"`
    Announce string `bencode:"announce"`
    AnnounceList [][]string `bencode:"announce-list"`
    Encoding string `bencode:"encoding"`
}

type bencodeInfo struct {
    PieceLength int `bencode:"piece length"`
    Pieces []byte `bencode:"pieces"`
    Name string `bencode:"name"`
    Length int `bencode:"length"`
    Files []bencodeFile `bencode:"files"`
}

type bencodeFile struct {
    Length int `bencode:"length"`
    Path []string `bencode:"path"`
}

func (to *Torrent) setID() {
    rand.Seed(time.Now().UnixNano())
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    id := "--GT0100--"

    for i := 0; i < 10; i++ {
        pos := rand.Intn(len(chars))
        id += string(chars[pos])
    }

    for i, c := range id {
        to.ID[i] = byte(c)
    }
}

func (meta bencodeMeta) String() string {
    var result string
    result += "Name: " + meta.Info.Name + "\n"
    result += "Announce: " + meta.Announce + "\n"
    for _, addr := range meta.AnnounceList {
        result += "Announce: " + addr[0] + "\n"
    }
    result += "Encoding: " + meta.Encoding + "\n"
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

func (to Torrent) getMeta() (bencodeMeta, error) {
    var meta bencodeMeta
    file, err := os.Open(to.Filename)
    if err != nil {
        log.Println("Could not open file:", to.Filename)
        return meta, err
    }
    defer file.Close()

    err = bencode.Unmarshal(file, &meta)
    if err != nil {
        log.Println("Could not unmarshal bencoded file:", to.Filename)
        return meta, err
    }

    return meta, nil
}
