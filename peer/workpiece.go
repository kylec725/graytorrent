package peer

import (
	"time"

	"github.com/kylec725/graytorrent/common"
	"github.com/pkg/errors"
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
	p.queue = append(p.queue, newWork)
}

func (p *Peer) removeWorkPiece(index int) {
	for i, workPiece := range p.queue {
		if index == workPiece.index {
			p.queue[i] = p.queue[len(p.queue)-1]
			p.queue = p.queue[:len(p.queue)-1]
			return
		}
	}
}

// requestAll starts requesting pieces for a peer once we are unchoked
func (p *Peer) requestAll() error {
	var err error
	for _, wp := range p.queue {
		if err = p.nextBlock(wp.index); err != nil {
			return errors.Wrap(err, "requestAll")
		}
	}
	return nil
}

// clearWork sends peer's work back into the work pool
func (p *Peer) clearWork(work chan int) {
	for _, wp := range p.queue {
		work <- wp.index
	}
	p.queue = p.queue[:0]
}
