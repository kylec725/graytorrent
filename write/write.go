package write

import (
    "os"
    "path/filepath"

    "github.com/kylec725/graytorrent/torrent"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrFileExists = errors.New("Torrent's file already exists")
)

// NewWrite sets up a new torrent file to write to
func NewWrite(to torrent.Torrent) error {
    for _, path := range to.Paths {
        // Return an error if the file already exists
        if _, err := os.Stat(path.Path); err == nil {
            return errors.Wrapf(ErrFileExists, "NewWrite %s", path.Path)
        }

        // Create directories recursively if necessary
        if dir := filepath.Dir(path.Path); dir != "" {
            err := os.MkdirAll(dir, 0755)
            if err != nil {
                return errors.Wrap(err, "NewWrite")
            }
        }

        _, err := os.Create(path.Path)
        if err != nil {
            return errors.Wrap(err, "NewWrite")
        }
    }

    return nil
}

func testFunc(to torrent.Torrent) int {
    return to.PieceLength
}
