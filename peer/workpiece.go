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
	size      int // size of the piece
	startTime time.Time
}

func (p *Peer) addWorkPiece(info common.TorrentInfo, index int) {
	pieceSize := common.PieceSize(info, index)
	piece := make([]byte, pieceSize)
	p.workPieces[index] = workPiece{index, piece, pieceSize, 0, pieceSize, time.Now()}
}

// clearWork sends peer's work back into the work pool
func (p *Peer) clearWork(work chan int) {
	for _, wp := range p.workPieces {
		work <- wp.index
	}
	p.workPieces = make(map[int]workPiece)
}
