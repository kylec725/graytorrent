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

const reqSize = 16384 // 16 kilobytes
const kb = 1024
const startQueue = 5 // Uses adaptive rate after first requests
const maxQueue = 625 // Maximum number of requests that can be sent out
const adjustTime = 5 // How often in seconds to adjust the queuing rate

// Errors
var (
	ErrBitfield  = errors.New("Malformed bitfield received")
	ErrMessage   = errors.New("Malformed message received")
	ErrPieceHash = errors.New("Received piece with bad hash")
)

func (p *Peer) sendMessage(msg *message.Message) error {
	if msg != nil {
		switch msg.ID {
		case message.MsgChoke:
			p.AmChoking = true
		case message.MsgUnchoke:
			p.AmChoking = false
		}
	}
	_, err := p.Conn.Write(msg.Encode())
	p.lastMsgSent = time.Now()
	return errors.Wrap(err, "sendMessage")
}

func (p *Peer) handleMessage(msg *message.Message, info common.TorrentInfo, work chan int, results chan int) error {
	if msg == nil {
		return nil // keep-alive message
	}
	switch msg.ID {
	case message.MsgChoke:
		p.PeerChoking = true
		p.clearWork(work) // Send back our work if we get choked
	case message.MsgUnchoke:
		p.PeerChoking = false
		err := p.requestAll() // Request pieces in our queue once we get unchoked
		return errors.Wrap(err, "handleMessage")
	case message.MsgInterested:
		p.PeerInterested = true
	case message.MsgNotInterested:
		p.PeerInterested = false
	case message.MsgHave: // TODO: use one case for checking for expected payload size
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

// TODO: limit the client's upload rate, may need to queue requests that come in
func (p *Peer) handleRequest(msg *message.Message, info common.TorrentInfo) error {
	if p.AmChoking { // Ignore requests if we are choking
		return nil
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

	// Update peer's amount downloaded
	p.bytesRcvd += uint32(len(block))
	go func() { // Only keep track of download rate within the adjustTime
		time.Sleep(adjustTime * time.Second)
		p.bytesRcvd -= uint32(len(block))
	}()

	// If piece is not in work queue, nothing happens
	for i := range p.queue { // We want to operate directly on the queue pieces
		if index == uint32(p.queue[i].index) {
			p.queue[i].left -= len(block)
			p.lastRequest = time.Now()

			err := write.AddBlock(info, int(index), int(begin), block, p.queue[i].piece)
			if err != nil {
				errors.Wrap(err, "handlePiece")
			}
			// If piece isn't done, request next piece and exit
			if p.queue[i].left > 0 {
				err := p.nextBlock(int(index))
				return errors.Wrap(err, "handlePiece")
			}

			// Piece is done: Verify hash then write
			if !write.VerifyPiece(info, int(index), p.queue[i].piece) { // Return to work pool if hash is incorrect
				p.removeWorkPiece(int(index))
				return errors.Wrap(ErrPieceHash, "handlePiece")
			}
			if err = write.AddPiece(info, int(index), p.queue[i].piece); err != nil { // Write piece to file
				log.WithFields(log.Fields{"peer": p.String(), "piece index": index, "error": err.Error()}).Debug("Writing piece to file failed")
				p.removeWorkPiece(int(index))
				return errors.Wrap(err, "handlePiece")
			}
			log.WithFields(log.Fields{"peer": p.String(), "piece index": index, "Rate": p.Rate()}).Trace("Wrote piece to file")

			// Write was successful
			p.removeWorkPiece(int(index))
			results <- int(index) // Notify main that a piece is done

			// Send not interested if necessary
			if len(p.queue) == 0 {
				msg := message.NotInterested()
				if _, err := p.Conn.Write(msg.Encode()); err != nil {
					return errors.Wrap(err, "downloadPiece")
				}
				p.AmInterested = false
			}
			break // Exit loop early on successful write
		}
	}
	return nil
}

// nextBlock requests the next block in a piece
func (p *Peer) nextBlock(index int) error {
	for i := range p.queue {
		if index == p.queue[i].index {
			length := common.Min(p.queue[i].left, reqSize)
			msg := message.Request(uint32(index), uint32(p.queue[i].curr), uint32(length))
			err := p.sendMessage(&msg)
			p.queue[i].curr += length
			p.lastRequest = time.Now()
			return errors.Wrap(err, "nextBlock")
		}
	}
	return nil
}

// downloadPiece starts a routine to download a piece from a peer
func (p *Peer) downloadPiece(info common.TorrentInfo, index int) error {
	if !p.AmInterested {
		msg := message.Interested()
		if _, err := p.Conn.Write(msg.Encode()); err != nil {
			return errors.Wrap(err, "downloadPiece")
		}
		p.AmInterested = true
	}
	p.addWorkPiece(info, index)
	if !p.PeerChoking { // If peer is choking, we call nextBlock when we receive an unchoke
		err := p.nextBlock(index)
		return errors.Wrap(err, "downloadPiece")
	}
	return nil
}

// Rate returns the current download rate in bytes/sec
func (p *Peer) Rate() uint32 {
	return p.bytesRcvd / adjustTime
}

// adjustRate changes the amount of requests to send out based on the download speed
func (p *Peer) adjustRate() {
	currRate := int(p.Rate() / kb) // we use rate in kb/sec for calculating the new rate

	// Use aggressive algorithm from rtorrent
	if currRate < 20 {
		p.maxQueue = currRate + 2
	} else {
		p.maxQueue = currRate/5 + 18
	}
	if p.maxQueue > maxQueue {
		p.maxQueue = maxQueue
	}
}
