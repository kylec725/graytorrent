package main

import (
    "fmt"
    "time"
    "context"

    "github.com/kylec725/graytorrent/torrent"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

// Command provides main with facilities to manage its torrents

func addTorrent(ctx context.Context, filename string) (torrent.Torrent, error) {
    to := torrent.Torrent{Path: filename}
    if err := to.Setup(ctx); err != nil {
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

func saveTorrents() {
    for i := range torrentList {
        torrentList[i].Save()
    }
}

func singleTorrent(ctx context.Context) {
    to, err := addTorrent(ctx, filename)
    if err != nil {
        fmt.Println("Single torrent failed:", err)
        log.WithFields(log.Fields{"filename": filename, "error": err.Error()}).Info("Failed to add torrent")
        return
    }
    log.WithField("name", to.Info.Name).Info("Torrent added")
    go to.Start(ctx)
    for to.Info.Left > 0 {
        time.Sleep(time.Second)
    }
    to.Save()
    fmt.Println("Torrent done:", to.Info.Name)
}
