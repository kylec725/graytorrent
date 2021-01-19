package write

import (
    "os"
    "path/filepath"
    // "io/ioutil"
    "fmt"

    "github.com/kylec725/graytorrent/torrent"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrFileExists = errors.New("Torrent's file already exists")
    ErrBadBlockBounds = errors.New("Received invalid bounds for a block")
    ErrCopyFailed = errors.New("Unexpected number of bytes copied")
    ErrInvalidPiece = errors.New("Piece index was out of bounds")
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

func pieceSize(to *torrent.Torrent, index int) int {
    if index == to.TotalPieces - 1 {
        return to.TotalLength - (to.TotalPieces - 1) * to.PieceLength
    }
    return to.PieceLength
}

// filesInPiece returns the indexes of files the piece is a part of
func filesInPiece(to *torrent.Torrent, index int) []int {
    var filesInPiece []int
    start := index * to.PieceLength // start byte index
    end := start + to.PieceLength // end byte index + 1
    if end > to.TotalLength {
        end = to.TotalLength
    }

    // Add all files within the piece's range to its list
    curr := 0
    for i, path := range to.Paths {
        curr += path.Length
        // Any file after the piece's start is part of the piece
        if curr > start {
            filesInPiece = append(filesInPiece, i)
        }
        // Exit once we past the last byte in the piece
        if curr >= end {
            break
        }
    }
    return filesInPiece
}

// AddBlock adds a block to a piece
func AddBlock(to *torrent.Torrent, index, begin int, block, piece []byte) error {
    if index < 0 || index >= to.TotalPieces {
        return errors.Wrap(ErrInvalidPiece, "AddBlock")
    }
    pieceSize := pieceSize(to, index)
    end := begin + len(block) // last index + 1 in the block

    // Check if bounds are possible or if integer overflow has occurred
    if begin < 0 || begin > (pieceSize - 1) || end - 1 < 0 || end > pieceSize {
        return errors.Wrap(ErrBadBlockBounds, "AddBlock")
    }

    bytesCopied := copy(piece[begin:end], block)
    if bytesCopied != len(block) {
        return errors.Wrap(ErrCopyFailed, "AddBlock")
    }

    return nil
}

// AddPiece takes a torrent piece, and writes it to the appropriate file
func AddPiece(to *torrent.Torrent, index int, piece []byte) error {
    if index < 0 || index >= to.TotalPieces {
        return errors.Wrap(ErrInvalidPiece, "AddPiece")
    }
    filesList := filesInPiece(to, index)

    for _, i := range filesList {
        fmt.Println("file:", to.Paths[i].Path)
    }

    return nil
}

// GetPiece returns a piece of a torrent as a byte slice
func GetPiece(to *torrent.Torrent, index int) ([]byte, error) {
    if index < 0 || index >= to.TotalPieces {
        return nil, errors.Wrap(ErrInvalidPiece, "GetPiece")
    }

    return nil, nil
}

// VerifyPiece checks that a completed piece has the correct hash
func VerifyPiece(to *torrent.Torrent, index int, piece []byte) bool {
    return false
}
