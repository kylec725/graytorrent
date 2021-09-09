package torrent

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

// Save saves data about a managed torrent's state to a file
func (to *Torrent) Save() error {
	// TODO: log results of saving
	// TODO: consider have a directory, with a file for each torrent's state
	// TODO: alternative: open history file json maybe, see if we are in it, if not: add ourselves
	//      if we are already, update info
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
