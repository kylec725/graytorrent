package main

import (
    "github.com/kylec725/graytorrent/torrent"
    "github.com/pkg/errors"
)

/*
Command provides main with facilities to manage its
list of torrents.
*/

func addTorrent(torrentList []torrent.Torrent, filename string, port uint16) (torrent.Torrent, error) {
    to := torrent.Torrent{Path: filename, Port: port}
    if err := to.Setup(); err != nil {
        return torrent.Torrent{}, errors.Wrap(err, "addTorrent")
    }
    torrentList = append(torrentList, to)
    return to, nil
}

func startTorrent(to torrent.Torrent) {
    return
}

func stopTorrent(to torrent.Torrent) {
    return
}

func removeTorrent(to torrent.Torrent) {
    return
}

func shutdown(torrentList []torrent.Torrent) {
    for i := range torrentList {
        torrentList[i].Save()
    }
}
