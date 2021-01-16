package write

import (
    "os"

    "github.com/kylec725/graytorrent/torrent"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrFileExists = errors.New("Torrent's filename path already exists")
)

// NewWrite sets up a new torrent file to write to
func NewWrite(to torrent.Torrent) error {
    // Return an error if the file already exists
    if _, err := os.Stat(to.Name); err == nil {
        return errors.Wrap(ErrFileExists, "NewWrite")
    }

    _, err := os.Create(to.Name)
    if err != nil {
        return errors.Wrap(err, "NewWrite")
    }

    return nil
}

func testFunc(to torrent.Torrent) int {
    return to.PieceLength
}
