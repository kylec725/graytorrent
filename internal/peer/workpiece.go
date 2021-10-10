package peer

import (
	"github.com/kylec725/gray/internal/common"
)

type workPiece struct {
	piece []byte
	left  int // bytes remaining in piece
	curr  int // current byte position in slice
	size  int // size of the piece
}

func (p *Peer) addWorkPiece(info common.TorrentInfo, index int) {
	pieceSize := common.PieceSize(info, index)
	piece := make([]byte, pieceSize)
	p.workPieces[index] = workPiece{piece, pieceSize, 0, pieceSize}
}

// clearWork sends peer's work back into the work pool
func (p *Peer) clearWork(work chan int) {
	for index := range p.workPieces {
		work <- index
	}
	p.workPieces = make(map[int]workPiece)
}
