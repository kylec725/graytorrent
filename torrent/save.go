package torrent

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config/graytorrent")
)

func init() {
	newPath := filepath.Join(grayTorrentPath, ".torrents")
	err := os.MkdirAll(newPath, os.ModePerm)
	if err != nil {
		log.Fatal("Error creating directory for torrent save data")
	}
}

// Save saves data about a managed torrent's state to a file NOTE: may want to add a directory parameter
func (to *Torrent) Save() error {
	// NOTE: have directory and save each torrent as a separate json
	// NOTE: alternative: open history file json maybe, see if we are in it then update info or add ourselves
	jsonStream, err := json.Marshal(to)
	if err != nil {
		return errors.Wrap(err, "Save")
	}

	err = os.WriteFile(to.Info.Name+".json", jsonStream, 0644)
	if err != nil {
		return errors.Wrap(err, "Save")
	}

	return nil
}

// TODO: Load torrents in dir
