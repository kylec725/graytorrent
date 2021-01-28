package peer

import (
	"time"

	"github.com/kylec725/graytorrent/common"
)

type workPiece struct {
	index     int
	piece     []byte
	left      int // bytes remaining in piece
	curr      int // current byte position in slice
	startTime time.Time
}

func (p *Peer) addWorkPiece(info common.TorrentInfo, index int) {
	pieceSize := common.PieceSize(info, index)
	piece := make([]byte, pieceSize)
	newWork := workPiece{index, piece, pieceSize, 0, time.Now()}
	p.workQueue = append(p.workQueue, newWork)
}

func (p *Peer) removeWorkPiece(index int) {
	removeIndex := -1
	for i, workPiece := range p.workQueue {
		if index == workPiece.index {
			removeIndex = i
		}
	}
	if removeIndex != -1 {
		p.workQueue[removeIndex] = p.workQueue[len(p.workQueue)-1]
		p.workQueue = p.workQueue[:len(p.workQueue)-1]
	}
}

// clearWork sends peer's work back into the work pool
func (p *Peer) clearWork(work chan int) {
	for _, wp := range p.workQueue {
		work <- wp.index
	}
	p.workQueue = p.workQueue[:0]
}
