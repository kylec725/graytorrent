package peer

import (
    "encoding/binary"
    "math"
    "time"
    "fmt"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/peer/message"
    "github.com/kylec725/graytorrent/write"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const reqSize = 16384  // kilobyte

// Errors
var (
    ErrBitfield = errors.New("Malformed bitfield received")
    ErrMessage = errors.New("Malformed message received")
    ErrPieceHash = errors.New("Received piece with bad hash")
)

func (peer *Peer) handleMessage(msg *message.Message, work chan int, results chan bool) error {
    if msg == nil {
        return nil  // keep-alive message
    }
    switch msg.ID {
    case message.MsgChoke:
        peer.peerChoking = true
    case message.MsgUnchoke:
        peer.peerChoking = false
    case message.MsgInterested:
        peer.peerInterested = true
    case message.MsgNotInterested:
        peer.peerInterested = false
    case message.MsgHave:
        if len(msg.Payload) != 4 {
            return errors.Wrap(ErrMessage, "handleMessage")
        }
        index := binary.BigEndian.Uint32(msg.Payload)
        peer.bitfield.Set(int(index))
    case message.MsgBitfield:
        expected := int(math.Ceil(float64(peer.info.TotalPieces) / 8))
        if len(msg.Payload) != expected {
            return errors.Wrap(ErrBitfield, "handleMessage")
        }
        peer.bitfield = msg.Payload
    case message.MsgRequest:
        if len(msg.Payload) != 12 {
            return errors.Wrap(ErrMessage, "handleMessage")
        }
        err := peer.handleRequest(msg)
        return errors.Wrap(err, "handleMessage")
    case message.MsgPiece:
        if len(msg.Payload) < 9 {
            return errors.Wrap(ErrMessage, "handleMessage")
        }
        err := peer.handlePiece(msg, work, results)
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
func (peer *Peer) sendRequest(index, begin, length int) error {
    msg := message.Request(uint32(index), uint32(begin), uint32(length))
    err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "sendRequest")
}

func (peer *Peer) handleRequest(msg *message.Message) error {
    if peer.amChoking {  // Tell the peer we are choking them and return
        chokeMsg := message.Choke()
        err := peer.Conn.Write(chokeMsg.Encode())
        return errors.Wrap(err, "handleRequest")
    }

    index := binary.BigEndian.Uint32(msg.Payload[0:4])
    begin := binary.BigEndian.Uint32(msg.Payload[4:8])
    length := binary.BigEndian.Uint32(msg.Payload[8:12])
    if !peer.info.Bitfield.Has(int(index)) {  // Ignore request if we don't have the piece
        return nil
    }

    piece, err := write.ReadPiece(peer.info, int(index))
    if err != nil {
        return errors.Wrap(err, "handleRequest")
    } else if len(piece) < int(begin + length) {  // Ignore request if the bounds aren't possible
        return nil
    }
    pieceMsg := message.Piece(index, begin, piece[begin:begin+length])
    err = peer.Conn.Write(pieceMsg.Encode())
    return errors.Wrap(err, "handleRequest")
}

// handlePiece adds a block to a piece we are getting
func (peer *Peer) handlePiece(msg *message.Message, work chan int, results chan bool) error {
    index := binary.BigEndian.Uint32(msg.Payload[0:4])
    begin := binary.BigEndian.Uint32(msg.Payload[4:8])
    block := msg.Payload[8:]

    // If piece is not in work queue, nothing happens
    for i := range peer.workQueue {  // We want to operate directly on the workQueue pieces
        if index == uint32(peer.workQueue[i].index) {
            peer.workQueue[i].left -= len(block)
            err := write.AddBlock(peer.info, int(index), int(begin), block, peer.workQueue[i].piece)
            if err != nil {
                errors.Wrap(err, "handlePiece")
            }
            // If piece isn't done, request next piece and exit
            if peer.workQueue[i].left > 0 {
                err := peer.nextBlock(int(index))
                return errors.Wrap(err, "handlePiece")
            }

            // Piece is done: Verify hash then write
            peer.adjustRate(peer.workQueue[i])  // Change rate regardless whether piece was correct
            if !write.VerifyPiece(peer.info, int(index), peer.workQueue[i].piece) {  // Return to work pool if hash is incorrect
                work <- int(index)
                peer.removeWorkPiece(int(index))
                return errors.Wrap(ErrPieceHash, "handlePiece")
            }
            if err = write.AddPiece(peer.info, int(index), peer.workQueue[i].piece); err != nil {  // Write piece to file
                log.WithFields(log.Fields{"peer": peer.String(), "piece index": index, "error": err.Error()}).Debug("Writing piece to file failed")
                work <- int(index)
                peer.removeWorkPiece(int(index))
                return errors.Wrap(err, "handlePiece")
            }
            log.WithFields(log.Fields{"peer": peer.String(), "piece index": index, "rate": peer.rate}).Trace("Wrote piece to file")

            // Write was successful
            peer.info.Left -= peer.workQueue[i].curr
            peer.removeWorkPiece(int(index))
            peer.info.Bitfield.Set(int(index))
            results <- true  // Notify main that a piece is done

            // Send not interested if necessary
            if len(peer.workQueue) == 0 {
                msg := message.NotInterested()
                if err := peer.Conn.Write(msg.Encode()); err != nil {
                    return errors.Wrap(err, "downloadPiece")
                }
                peer.amInterested = false
            }
            break  // Exit loop early on successful write
        }
    }
    return nil
}

// nextBlock requests the next block in a piece
func (peer *Peer) nextBlock(index int) error {
    for i := range peer.workQueue {
        if index == peer.workQueue[i].index {
            length := common.Min(peer.workQueue[i].left, reqSize)
            err := peer.sendRequest(index, peer.workQueue[i].curr, length)
            peer.workQueue[i].curr += length
            return errors.Wrap(err, "nextBlock")
        }
    }
    return nil
}

// downloadPiece starts a routine to download a piece from a peer
func (peer *Peer) downloadPiece(index int) error {
    if !peer.amInterested {
        msg := message.Interested()
        if err := peer.Conn.Write(msg.Encode()); err != nil {
            return errors.Wrap(err, "downloadPiece")
        }
        peer.amInterested = true
    }
    peer.addWorkPiece(index)
    err := peer.nextBlock(index)  // First block request
    return errors.Wrap(err, "downloadPiece")
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
