package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kylec725/graytorrent/torrent"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Manager provides main with facilities to manage its torrents

func addTorrent(ctx context.Context, filename string) (torrent.Torrent, error) {
	to := torrent.Torrent{Path: filename}
	if err := to.Setup(ctx); err != nil {
		return torrent.Torrent{}, errors.Wrap(err, "addTorrent")
	}
	torrentList = append(torrentList, to)
	return to, nil
}

func removeTorrent(to torrent.Torrent) {
	to.Cancel()
	for i := range torrentList {
		if to.Path == torrentList[i].Path {
			torrentList[i] = torrentList[len(torrentList)-1]
			torrentList = torrentList[:len(torrentList)-1]
			return
		}
	}
}

func startTorrent(to torrent.Torrent) {
	go to.Start()
}

func stopTorrent(to torrent.Torrent) {
	to.Cancel()
}

func singleTorrent(ctx context.Context) {
	to, err := addTorrent(ctx, filename)
	if err != nil {
		fmt.Println("Single torrent failed:", err)
		log.WithFields(log.Fields{"filename": filename, "error": err.Error()}).Info("Failed to add torrent")
		return
	}
	log.WithField("name", to.Info.Name).Info("Torrent added")
	go to.Start()
	for to.Info.Left > 0 { // Go compiler marks this as data race, not a big deal, we're just polling the value
		time.Sleep(time.Second)
	}
	torrent.SaveAll(torrentList)
	fmt.Println("Torrent done:", to.Info.Name)
}
