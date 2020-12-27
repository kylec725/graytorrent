package torrent

import (
    "log"
    "os"
    "strconv"

    bencode "github.com/jackpal/bencode-go"
)

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
