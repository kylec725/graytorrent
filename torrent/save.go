package torrent

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const saveDataType = ".gz"

var (
	grayPath     = filepath.Join(os.Getenv("HOME"), ".config", "gray")
	saveDataPath = filepath.Join(grayPath, ".torrents"+saveDataType)
)

// SaveAll saves the states of all managed torrents
func SaveAll(torrentList []Torrent) error {
	jsonStream, err := json.Marshal(torrentList)
	if err != nil {
		return errors.Wrap(err, "SaveAll")
	}

	file, err := os.Create(saveDataPath) // os.Create creates or truncates the named file
	if err != nil {
		return errors.Wrap(err, "SaveAll")
	}
	defer file.Close()

	// Write the file using gzip compression
	writer := gzip.NewWriter(file)
	defer writer.Close()
	_, err = writer.Write(jsonStream)
	if err != nil {
		return errors.Wrap(err, "SaveAll")
	}

	return nil
}

// LoadAll retrieves a list of all managed torrents
func LoadAll() ([]Torrent, error) {
	var torrentList []Torrent

	file, err := os.Open(saveDataPath)
	if err != nil { // If the save data file doesn't exist, return an empty list
		return torrentList, nil
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, errors.Wrap(err, "LoadAll")
	}
	defer reader.Close()

	fileBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "LoadAll")
	}

	err = json.Unmarshal(fileBytes, &torrentList)
	if err != nil {
		return nil, errors.Wrap(err, "LoadAll")
	}

	return torrentList, nil
}

// TODO: verify existing file's present pieces
