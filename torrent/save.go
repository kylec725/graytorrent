package torrent

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	torrentDataPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent", ".torrents")
)

func init() {
	err := os.MkdirAll(torrentDataPath, os.ModePerm)
	if err != nil {
		log.Fatal("Error creating directory for torrent save data")
	}
}

// dataFile returns a path to the torrent's GrayTorrent data file
func (to *Torrent) dataFile() string {
	return filepath.Join(torrentDataPath, to.Info.Name+".json")
}

// Save saves data about a managed torrent's state to a file
func (to *Torrent) Save() error {
	// NOTE: alternative: open history file json, see if we are in it, then update info or add ourselves
	jsonStream, err := json.Marshal(to)
	if err != nil {
		return errors.Wrap(err, "Save")
	}

	file, err := os.Create(to.dataFile()) // os.Create creates or truncates the named file
	if err != nil {
		return errors.Wrap(err, "Save")
	}

	_, err = file.Write(jsonStream)
	if err != nil {
		return errors.Wrap(err, "Save")
	}

	return nil
}

// TODO: Load torrents in dir
