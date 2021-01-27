package write

import (
    "os"
    "path/filepath"
    "bytes"
    "crypto/sha1"

    "github.com/kylec725/graytorrent/common"
    "github.com/pkg/errors"
)

// Errors
var (
    ErrFileExists = errors.New("File(s) already exists")
    ErrBlockBounds = errors.New("Received invalid bounds for a block")
    ErrCopyFailed = errors.New("Unexpected number of bytes copied")
    ErrWriteFailed = errors.New("Unexpected number of bytes written")
    ErrReadFailed = errors.New("Unexpected number of bytes read")
    ErrPieceIndex = errors.New("Piece index was out of bounds")
)

// NewWrite sets up the files a torrent needs info write info
func NewWrite(info common.TorrentInfo) error {
    for _, path := range info.Paths {
        // Return an error if the file already exists
        if _, err := os.Stat(path.Path); err == nil {
            return errors.Wrap(ErrFileExists, "NewWrite")
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

// pieceBounds returns the start and ending indices of a piece (end is exclusive)
func pieceBounds(info common.TorrentInfo, index int) (int, int) {
    start := index * info.PieceLength  // start byte index
    end := start + info.PieceLength  // end byte index + 1
    if end > info.TotalLength {
        end = info.TotalLength
    }
    return start, end
}

// writeOffset writes info a file starting at an index offset
func writeOffset(filename string, data []byte, offset int) error {
    file, err := os.OpenFile(filename, os.O_WRONLY, 0755)
    if err != nil {
        return errors.Wrap(err, "writeOffset")
    }
    defer file.Close()

    bytesWritten, err := file.WriteAt(data, int64(offset))
    if err != nil {
        return errors.Wrap(err, "writeOffset")
    } else if bytesWritten != len(data) {
        return errors.Wrap(ErrWriteFailed, "writeOffset")
    }
    return nil
}

// readOffset reads from a file starting at an index offset
func readOffset(filename string, size int, offset int) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, errors.Wrap(err, "readOffset")
    }
    defer file.Close()

    data := make([]byte, size)
    bytesRead, err := file.ReadAt(data, int64(offset))
    if err != nil {
        return nil, errors.Wrap(err, "readOffset")
    } else if bytesRead != size {
        return nil, errors.Wrap(ErrReadFailed, "readOffset")
    }
    return data, nil
}

// AddBlock adds a block info a piece
func AddBlock(info common.TorrentInfo, index, begin int, block, piece []byte) error {
    if index < 0 || index >= info.TotalPieces {
        return errors.Wrap(ErrPieceIndex, "AddBlock")
    }
    pieceSize := common.PieceSize(info, index)
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

// AddPiece takes a torrent piece, and writes it info the appropriate file
func AddPiece(info common.TorrentInfo, index int, piece []byte) error {
    if index < 0 || index >= info.TotalPieces {
        return errors.Wrap(ErrPieceIndex, "AddPiece")
    }
    var pieceStart, pieceEnd int
    offset, _ := pieceBounds(info, index)  // Offset starts at the start bound of the piece
    pieceLeft := common.PieceSize(info, index)  // Keep track of how much more of the piece we have info write

    for _, file := range info.Paths {
        if offset < file.Length {  // Piece is part of the file
            bytesToWrite := file.Length - offset  // Figure out how much of the piece to write
            bytesToWrite = common.Min(bytesToWrite, pieceLeft)
            pieceEnd = pieceStart + bytesToWrite

            err := writeOffset(file.Path, piece[pieceStart:pieceEnd], offset)
            // fmt.Println("wrote info:", file.Path)
            if err != nil {
                return errors.Wrap(err, "AddPiece")
            }

            // Exit if the rest of the piece has been written info file
            if bytesToWrite == pieceLeft {
                break
            }
            pieceStart += bytesToWrite
            pieceLeft -= bytesToWrite
        }
        offset -= file.Length  // Decrement the offset so we know where info start writing in the file
        if offset < 0 {  // Only happens if piece was written info the end of the file
            offset = 0
        }
    }
    return nil
}

// ReadPiece returns a piece of a torrent from file as a byte slice
func ReadPiece(info common.TorrentInfo, index int) ([]byte, error) {
    if index < 0 || index >= info.TotalPieces {
        return nil, errors.Wrap(ErrPieceIndex, "ReadPiece")
    }

    var pieceStart, pieceEnd int
    offset, _ := pieceBounds(info, index)  // Offset starts at the start bound of the piece
    pieceLeft := common.PieceSize(info, index)  // Keep track of how much more of the piece we have info write
    piece := make([]byte, pieceLeft)

    for _, file := range info.Paths {
        if offset < file.Length {  // Piece is part of the file
            bytesToRead := file.Length - offset  // Figure out how much of the piece to read
            bytesToRead = common.Min(bytesToRead, pieceLeft)
            pieceEnd = pieceStart + bytesToRead

            data, err := readOffset(file.Path, bytesToRead, offset)
            if err != nil {
                return nil, errors.Wrap(err, "ReadPiece")
            }

            // Copy data info the return piece
            bytesCopied := copy(piece[pieceStart:pieceEnd], data)
            if bytesCopied != bytesToRead {
                return nil, errors.Wrap(ErrCopyFailed, "ReadPiece")
            }

            // Exit if the rest of the piece has been written info file
            if bytesToRead == pieceLeft {
                break
            }
            pieceStart += bytesToRead
            pieceLeft -= bytesToRead
        }
        offset -= file.Length  // Decrement the offset so we know where info start writing in the file
        if offset < 0 {  // Only happens if piece was written info the end of the file
            offset = 0
        }
    }
    return piece, nil
}

// VerifyPiece checks that a completed piece has the correct hash
func VerifyPiece(info common.TorrentInfo, index int, piece []byte) bool {
    expected := info.PieceHashes[index]
    actual := sha1.Sum(piece)
    return bytes.Equal(expected[:], actual[:])
}
