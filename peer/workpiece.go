package peer

import (
    "time"

    "github.com/kylec725/graytorrent/common"
)

type workPiece struct {
    index int
    piece []byte
    left int  // bytes remaining in piece
    curr int  // current byte position in slice
    startTime time.Time
}

func (peer *Peer) addWorkPiece(index int) {
    pieceSize := common.PieceSize(peer.info, index)
    piece := make([]byte, pieceSize)
    newWork := workPiece{index, piece, pieceSize, 0, time.Now()}
    peer.workQueue = append(peer.workQueue, newWork)
}

func (peer *Peer) removeWorkPiece(index int) {
    removeIndex := -1
    for i, workPiece := range peer.workQueue {
        if index == workPiece.index {
            removeIndex = i
        }
    }
    if removeIndex != -1 {
        peer.workQueue[removeIndex] = peer.workQueue[len(peer.workQueue) - 1]
        peer.workQueue = peer.workQueue[:len(peer.workQueue) - 1]
    }
}

// adjustRate changes rate according to the work rate of reqSize per second
func (peer *Peer) adjustRate(wp workPiece) {
    duration := time.Since(wp.startTime)
    numBlocks := wp.curr / reqSize  // truncate number of blocks to be conservative
    currRate := float64(numBlocks) / duration.Seconds()  // reqSize per second

    // Use aggressive algorithm from rtorrent
    // if currRate < 20 {
    //     peer.rate = int(currRate) + 2
    // } else {
    //     peer.rate = int(currRate / 5 + 18)
    // }
    if currRate > 2 {
        peer.rate = int(currRate)
    } else {
        peer.rate = 2
    }
}
