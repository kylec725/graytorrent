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

const blockSize = 16384 // 16 kilobytes
const kb = 1024
const startQueue = 5 // Uses adaptive rate after first requests
const maxQueue = 625 // Maximum number of requests that can be sent out
const adjustTime = 5 // How often in seconds to adjust the transfer rates

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
		err := p.handlePiece(msg, info, work, results)
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

// TODO: limit the client's upload rate, may need to workPieces requests that come in
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

	// Update peer's amount uploaded
	p.bytesSent += uint32(length)
	go func() { // Only keep track of download rate within the adjustTime
		time.Sleep(adjustTime * time.Second)
		p.bytesSent -= uint32(length)
	}()

	return errors.Wrap(err, "handleRequest")
}

// handlePiece adds a block to a piece we are getting
func (p *Peer) handlePiece(msg *message.Message, info common.TorrentInfo, work chan int, results chan int) error {
	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	begin := binary.BigEndian.Uint32(msg.Payload[4:8])
	block := msg.Payload[8:]

	// Update peer's amount downloaded
	p.bytesRcvd += uint32(len(block))
	go func() { // Only keep track of download rate within the adjustTime
		time.Sleep(adjustTime * time.Second)
		p.bytesRcvd -= uint32(len(block))
	}()

	// If piece is not in workPieces, nothing happens
	if wp, ok := p.workPieces[int(index)]; ok {
		p.lastRequest = time.Now()
		p.queue--

		// Update the workpiece
		wp.left -= len(block)
		err := write.AddBlock(info, int(index), int(begin), block, wp.piece)
		if err != nil {
			errors.Wrap(err, "handlePiece")
		}
		p.workPieces[int(index)] = wp

		// If piece isn't done, request next piece and exit
		if p.workPieces[int(index)].left > 0 {
			// err := p.nextBlock(int(index))
			// return errors.Wrap(err, "handlePiece")
			return nil
		}

		// Piece is done: Verify hash then write
		if !write.VerifyPiece(info, int(index), p.workPieces[int(index)].piece) { // Return to work pool if hash is incorrect
			delete(p.workPieces, int(index))
			work <- int(index)
			return errors.Wrap(ErrPieceHash, "handlePiece")
		}
		if err = write.AddPiece(info, int(index), p.workPieces[int(index)].piece); err != nil { // Write piece to file
			delete(p.workPieces, int(index))
			work <- int(index)
			return errors.Wrap(err, "handlePiece")
		}
		log.WithFields(log.Fields{"peer": p.String(), "piece index": index, "DownRate": p.DownRatePretty()}).Trace("Wrote piece to file")

		// Write was successful
		delete(p.workPieces, int(index))
		results <- int(index) // Notify main that a piece is done

		// Send not interested if necessary
		if len(p.workPieces) == 0 {
			msg := message.NotInterested()
			if _, err := p.Conn.Write(msg.Encode()); err != nil {
				return errors.Wrap(err, "downloadPiece")
			}
			p.AmInterested = false
		}
	}
	return nil
}

// nextBlock requests the next block in a piece
func (p *Peer) nextBlock(index int) error {
	if wp, ok := p.workPieces[index]; ok {
		length := common.Min(wp.left, blockSize)
		msg := message.Request(uint32(index), uint32(wp.curr), uint32(length)) // Message must be sent before updating the workpiece curr value
		err := p.sendMessage(&msg)
		err = errors.WithMessagef(err, "index %d begin %d length %d", index, wp.curr, length)

		wp.curr += length
		p.workPieces[index] = wp
		p.queue++
		p.lastRequest = time.Now()
		return errors.Wrap(err, "nextBlock")
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
	return nil
}

// fillQueue sends out as many requests for blocks as the peer's queue allows
func (p *Peer) fillQueue() error {
	for i := range p.workPieces {
		for p.queue < p.maxQueue && p.workPieces[i].curr < p.workPieces[i].size {
			err := p.nextBlock(i)
			if err != nil {
				return errors.Wrap(err, "fillQueue")
			}
		}
	}
	return nil
}

// DownRate returns the current download rate in bytes/sec
func (p *Peer) DownRate() uint32 {
	return p.bytesRcvd / adjustTime
}

// DownRatePretty returns a human readable form of the download rate
func (p *Peer) DownRatePretty() string {
	rate := float64(p.DownRate())
	suffix := "B/s"
	if rate > 1024 {
		rate /= 1024
		suffix = "KiB/s"
	}
	if rate > 1024 {
		rate /= 1024
		suffix = "MiB/s"
	}
	return fmt.Sprintf("%.2f "+suffix, rate)
}

// UpRate returns the current download rate in bytes/sec
func (p *Peer) UpRate() uint32 {
	return p.bytesRcvd / adjustTime
}

// UpRatePretty returns a human readable form of the download rate
func (p *Peer) UpRatePretty() string {
	rate := float64(p.UpRate())
	suffix := "B/s"
	if rate > 1024 {
		rate /= 1024
		suffix = "KiB/s"
	}
	if rate > 1024 {
		rate /= 1024
		suffix = "MiB/s"
	}
	return fmt.Sprintf("%.2f "+suffix, rate)
}

// adjustRate changes the amount of requests to send out based on the download speed
func (p *Peer) adjustRate() {
	currRate := int(p.DownRate() / kb) // we use rate in kb/sec for calculating the new rate

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
