package peer

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/peer/message"
	"github.com/kylec725/graytorrent/write"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const reqSize = 16384 // kilobyte

// Errors
var (
	ErrBitfield  = errors.New("Malformed bitfield received")
	ErrMessage   = errors.New("Malformed message received")
	ErrPieceHash = errors.New("Received piece with bad hash")
)

func (p *Peer) handleMessage(msg *message.Message, info common.TorrentInfo, work chan int, results chan int) error {
	if msg == nil {
		return nil // keep-alive message
	}
	switch msg.ID {
	case message.MsgChoke:
		p.peerChoking = true
		p.clearWork(work)
	case message.MsgUnchoke:
		p.peerChoking = false
		err := p.requestAll()
		return errors.Wrap(err, "handleMessage")
	case message.MsgInterested:
		p.peerInterested = true
	case message.MsgNotInterested:
		p.peerInterested = false
	case message.MsgHave:
		if len(msg.Payload) != 4 {
			return errors.Wrap(ErrMessage, "handleMessage")
		}
		index := binary.BigEndian.Uint32(msg.Payload)
		p.bitfield.Set(int(index))
	case message.MsgBitfield:
		expected := int(math.Ceil(float64(info.TotalPieces) / 8))
		if len(msg.Payload) != expected {
			return errors.Wrap(ErrBitfield, "handleMessage")
		}
		p.bitfield = msg.Payload
	case message.MsgRequest:
		if len(msg.Payload) != 12 {
			return errors.Wrap(ErrMessage, "handleMessage")
		}
		err := p.handleRequest(msg, info)
		return errors.Wrap(err, "handleMessage")
	case message.MsgPiece:
		if len(msg.Payload) < 9 {
			return errors.Wrap(ErrMessage, "handleMessage")
		}
		err := p.handlePiece(msg, info, results)
		return errors.Wrap(err, "handleMessage")
	case message.MsgCancel:
		if len(msg.Payload) != 12 {
			return errors.Wrap(ErrMessage, "handleMessage")
		}
		fmt.Println("MsgCancel not yet implemented")
	case message.MsgPort:
		if len(msg.Payload) != 2 {
			return errors.Wrap(ErrMessage, "handleMessage")
		}
		fmt.Println("MsgPort not yet implemented")
	}
	return nil
}

// sendRequest sends a piece request message to a peer
func (p *Peer) sendRequest(index, begin, length int) error {
	msg := message.Request(uint32(index), uint32(begin), uint32(length))
	_, err := p.Conn.Write(msg.Encode())
	return errors.Wrap(err, "sendRequest")
}

func (p *Peer) handleRequest(msg *message.Message, info common.TorrentInfo) error {
	if p.amChoking { // Tell the peer we are choking them and return
		chokeMsg := message.Choke()
		_, err := p.Conn.Write(chokeMsg.Encode())
		return errors.Wrap(err, "handleRequest")
	}

	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	begin := binary.BigEndian.Uint32(msg.Payload[4:8])
	length := binary.BigEndian.Uint32(msg.Payload[8:12])
	if !info.Bitfield.Has(int(index)) { // Ignore request if we don't have the piece
		return nil
	}

	piece, err := write.ReadPiece(info, int(index))
	if err != nil {
		return errors.Wrap(err, "handleRequest")
	} else if len(piece) < int(begin+length) { // Ignore request if the bounds aren't possible
		return nil
	}
	pieceMsg := message.Piece(index, begin, piece[begin:begin+length])
	_, err = p.Conn.Write(pieceMsg.Encode())
	return errors.Wrap(err, "handleRequest")
}

// handlePiece adds a block to a piece we are getting
func (p *Peer) handlePiece(msg *message.Message, info common.TorrentInfo, results chan int) error {
	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	begin := binary.BigEndian.Uint32(msg.Payload[4:8])
	block := msg.Payload[8:]

	// If piece is not in work queue, nothing happens
	for i := range p.workQueue { // We want to operate directly on the workQueue pieces
		if index == uint32(p.workQueue[i].index) {
			p.workQueue[i].left -= len(block)
			p.lastRequest = time.Now()

			err := write.AddBlock(info, int(index), int(begin), block, p.workQueue[i].piece)
			if err != nil {
				errors.Wrap(err, "handlePiece")
			}
			// If piece isn't done, request next piece and exit
			if p.workQueue[i].left > 0 {
				err := p.nextBlock(int(index))
				return errors.Wrap(err, "handlePiece")
			}

			// Piece is done: Verify hash then write
			p.adjustRate(p.workQueue[i])                                    // Change rate regardless whether piece was correct
			if !write.VerifyPiece(info, int(index), p.workQueue[i].piece) { // Return to work pool if hash is incorrect
				p.removeWorkPiece(int(index))
				return errors.Wrap(ErrPieceHash, "handlePiece")
			}
			if err = write.AddPiece(info, int(index), p.workQueue[i].piece); err != nil { // Write piece to file
				log.WithFields(log.Fields{"peer": p.String(), "piece index": index, "error": err.Error()}).Debug("Writing piece to file failed")
				p.removeWorkPiece(int(index))
				return errors.Wrap(err, "handlePiece")
			}
			log.WithFields(log.Fields{"peer": p.String(), "piece index": index, "rate": p.rate}).Trace("Wrote piece to file")

			// Write was successful
			p.removeWorkPiece(int(index))
			results <- int(index) // Notify main that a piece is done

			// Send not interested if necessary
			if len(p.workQueue) == 0 {
				msg := message.NotInterested()
				if _, err := p.Conn.Write(msg.Encode()); err != nil {
					return errors.Wrap(err, "downloadPiece")
				}
				p.amInterested = false
			}
			break // Exit loop early on successful write
		}
	}
	return nil
}

// nextBlock requests the next block in a piece
func (p *Peer) nextBlock(index int) error {
	for i := range p.workQueue {
		if index == p.workQueue[i].index {
			length := common.Min(p.workQueue[i].left, reqSize)
			err := p.sendRequest(index, p.workQueue[i].curr, length)
			p.workQueue[i].curr += length
			p.lastRequest = time.Now()
			return errors.Wrap(err, "nextBlock")
		}
	}
	return nil
}

// downloadPiece starts a routine to download a piece from a peer
func (p *Peer) downloadPiece(info common.TorrentInfo, index int) error { // TODO make sure we are unchoked before sending requests
	if !p.amInterested {
		msg := message.Interested()
		if _, err := p.Conn.Write(msg.Encode()); err != nil {
			return errors.Wrap(err, "downloadPiece")
		}
		p.amInterested = true
	}
	p.addWorkPiece(info, index)
	return nil
}

// adjustRate changes rate according to the work rate of reqSize per second
func (p *Peer) adjustRate(wp workPiece) {
	duration := time.Since(wp.startTime)
	numBlocks := wp.curr / reqSize                      // truncate number of blocks to be conservative
	currRate := float64(numBlocks) / duration.Seconds() // reqSize per second

	// Use aggressive algorithm from rtorrent
	if currRate < 20 {
		p.rate = int(currRate) + 2
	} else {
		p.rate = int(currRate/5 + 18)
	}
}
