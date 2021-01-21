package write

import (
    "os"
    "path/filepath"
    "bytes"
    "crypto/sha1"

    "github.com/kylec725/graytorrent/torrent"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrFileExists = errors.New("Torrent's file already exists")
    ErrBlockBounds = errors.New("Received invalid bounds for a block")
    ErrCopyFailed = errors.New("Unexpected number of bytes copied")
    ErrPieceIndex = errors.New("Piece index was out of bounds")
)

// NewWrite sets up the files a torrent needs to write to
func NewWrite(to *torrent.Torrent) error {
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

// pieceBounds returns the start and ending indices of a piece (end is exclusive)
func pieceBounds(to *torrent.Torrent, index int) (int, int) {
    start := index * to.PieceLength  // start byte index
    end := start + to.PieceLength  // end byte index + 1
    if end > to.TotalLength {
        end = to.TotalLength
    }
    return start, end
}

// writeOffset writes to a file starting at an index offset
func writeOffset(filename string, data []byte, offset int) (int, error) {
    file, err := os.OpenFile(filename, os.O_WRONLY, 0755)
    if err != nil {
        return 0, errors.Wrap(err, "writeOffset")
    }
    defer file.Close()
    bytesWritten, err := file.WriteAt(data, int64(offset))
    if err != nil {
        return 0, errors.Wrap(err, "writeOffset")
    }
    return bytesWritten, nil
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

// // filesInPiece returns the indexes of files the piece is a part of
// func filesInPiece(to *torrent.Torrent, index int) []int {
//     var filesInPiece []int
//     start := index * to.PieceLength  // start byte index
//     end := start + to.PieceLength  // end byte index + 1
//     if end > to.TotalLength {
//         end = to.TotalLength
//     }
//
//     // Add all files within the piece's range to its list
//     curr := 0
//     for i, path := range to.Paths {
//         curr += path.Length
//         // Any file after the piece's start is part of the piece
//         if curr > start {
//             filesInPiece = append(filesInPiece, i)
//         }
//         // Exit once we past the last byte in the piece
//         if curr >= end {
//             break
//         }
//     }
//     return filesInPiece
// }

// AddBlock adds a block to a piece
func AddBlock(to *torrent.Torrent, index, begin int, block, piece []byte) error {
    if index < 0 || index >= to.TotalPieces {
        return errors.Wrap(ErrPieceIndex, "AddBlock")
    }
    pieceSize := pieceSize(to, index)
    end := begin + len(block)  // last index + 1 in the block

    // Check if bounds are possible or if integer overflow has occurred
    if begin < 0 || begin > (pieceSize - 1) || end - 1 < 0 || end > pieceSize {
        return errors.Wrap(ErrBlockBounds, "AddBlock")
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
        return errors.Wrap(ErrPieceIndex, "AddPiece")
    }
    var pieceStart, pieceEnd int
    offset, _ := pieceBounds(to, index)  // Offset starts at the start bound of the piece
    pieceLeft := pieceSize(to, index)  // Keep track of how much more of the piece we have to write

    for _, file := range to.Paths {
        if offset < file.Length {  // Piece is part of the file
            bytesToWrite := file.Length - offset  // Figure out how much of the piece to write to file
            bytesToWrite = min(bytesToWrite, pieceLeft)
            pieceEnd = pieceStart + bytesToWrite

            _, err := writeOffset(file.Path, piece[pieceStart:pieceEnd], offset)
            // fmt.Println("wrote to:", file.Path)
            if err != nil {
                return errors.Wrap(err, "AddPiece")
            }
            pieceStart += bytesToWrite

            // Exit if the rest of the piece has been written to file
            if bytesToWrite == pieceLeft {
                break
            }
            pieceLeft -= bytesToWrite
        }
        offset -= file.Length  // Decrement the offset so we know where to start writing in the file
        if offset < 0 {  // Only happens if piece was written to the end of the file
            offset = 0
        }
    }
    return nil
}

// GetPiece returns a piece of a torrent as a byte slice
func GetPiece(to *torrent.Torrent, index int) ([]byte, error) {
    if index < 0 || index >= to.TotalPieces {
        return nil, errors.Wrap(ErrPieceIndex, "GetPiece")
    }

    return nil, nil
}

// VerifyPiece checks that a completed piece has the correct hash
func VerifyPiece(to *torrent.Torrent, index int, piece []byte) bool {
    expected := to.PieceHashes[index]
    actual := sha1.Sum(piece)
    return bytes.Equal(expected[:], actual[:])
}
