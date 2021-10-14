package torrent

import (
	"compress/gzip"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/pkg/errors"
)

const saveType = ".gz"

func (to *Torrent) saveFile() string {
	return path.Join(common.SavePath, to.Info.Name+saveType)
}

// Save saves the progress of a torrent
func (to *Torrent) Save() error {
	// TODO: only save a file if the current last modified time matches our last modified time
	jsonStream, err := json.Marshal(to)
	if err != nil {
		return errors.Wrap(err, "Save")
	}

	file, err := os.Create(to.saveFile()) // os.Create creates or truncates the named file
	if err != nil {
		return errors.Wrap(err, "Save")
	}
	defer file.Close()

	// Write the file using gzip compression
	writer := gzip.NewWriter(file)
	defer writer.Close()
	_, err = writer.Write(jsonStream)
	return errors.Wrap(err, "Save")
}

// SaveAll saves the states of all managed torrents
func (s *Session) SaveAll() error {
	for _, to := range s.torrents {
		if err := to.Save(); err != nil {
			return err // No need to wrap err
		}
	}
	return nil
}

// LoadAll retrieves a list of all managed torrents
func LoadAll() (map[[20]byte]*Torrent, error) {
	torrents := make(map[[20]byte]*Torrent)

	err := filepath.WalkDir(common.SavePath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() { // Ignore directories
			return nil
		}
		file, err := os.Open(path)
		if err != nil { // If the save data file doesn't exist, return an empty list
			return err
		}
		defer file.Close()

		reader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer reader.Close()

		fileBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}

		var to Torrent
		err = json.Unmarshal(fileBytes, &to)
		if err != nil {
			return err
		}

		if err = to.Init(); err != nil {
			return err
		}

		torrents[to.Info.InfoHash] = &to

		return nil
	})

	return torrents, errors.Wrap(err, "LoadAll")
}

// TODO: verify existing file's present pieces
