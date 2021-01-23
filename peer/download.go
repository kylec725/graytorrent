package peer

import (
    "encoding/binary"
    "math"
    "fmt"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/peer/message"
    "github.com/kylec725/graytorrent/write"
    "github.com/pkg/errors"
)

const maxReqSize = 16384

// Errors
var (
    ErrBitfield = errors.New("Malformed bitfield received")
    ErrMessage = errors.New("Malformed message received")
    ErrPieceHash = errors.New("Received piece with bad hash")
    ErrUnexpectedPiece = errors.New("Received piece when not expecting it")
)

// getMessage reads in a message from the peer
func (peer *Peer) getMessage() (*message.Message, error) {
    buf := make([]byte, 4)
    if err := peer.Conn.Read(buf); err != nil {
        return nil, errors.Wrap(err, "getMessage")
    }
    msgLen := binary.BigEndian.Uint32(buf)
    if msgLen == 0 {  // Keep-alive message
        return nil, nil
    }

    buf = make([]byte, msgLen)
    err := peer.Conn.Read(buf)
    return message.Decode(buf), errors.Wrap(err, "getMessage")
}

func (peer *Peer) handleMessage(msg *message.Message, currentWork []byte) ([]byte, error) {
    if msg == nil {
        // reset keep-alive
        return currentWork, nil
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
        index := binary.BigEndian.Uint32(msg.Payload)
        peer.bitfield.Set(int(index))
    case message.MsgBitfield:
        expected := int(math.Ceil(float64(peer.info.TotalPieces) / 8))
        if expected != len(msg.Payload) {
            return currentWork, errors.Wrap(ErrBitfield, "handleMessage")
        }
        peer.bitfield = msg.Payload
    case message.MsgRequest:
        err := peer.handleRequest(msg)
        return currentWork, errors.Wrap(err, "handleMessage")
    case message.MsgPiece:
        currentWork, err := peer.handlePiece(msg, currentWork)
        return currentWork, errors.Wrap(err, "handleMessage")
    case message.MsgCancel:
        fmt.Println("MsgCancel not yet implemented")
    case message.MsgPort:
        fmt.Println("MsgPort not yet implemented")
    }
    return currentWork, nil
}

// sendRequest sends a piece request message to a peer
func (peer *Peer) sendRequest(index, begin, length int) error {
    msg := message.Request(index, begin, length)
    err := peer.Conn.Write(msg.Encode())
    return errors.Wrap(err, "sendRequest")
}

// TODO
func (peer *Peer) handleRequest(msg *message.Message) error {
    if peer.amChoking {  // Tell the peer we are choking them and return
        chokeMsg := message.Choke()
        err := peer.Conn.Write(chokeMsg.Encode())
        return errors.Wrap(err, "handleRequest")
    }
    return errors.New("Not yet implemented")
}

// handlePiece adds a MsgPiece to the current work slice
func (peer *Peer) handlePiece(msg *message.Message, currentWork []byte) ([]byte, error) {
    index := binary.BigEndian.Uint32(msg.Payload[0:4])
    begin := binary.BigEndian.Uint32(msg.Payload[4:8])
    block := msg.Payload[8:]

    err := write.AddBlock(peer.info, int(index), int(begin), block, currentWork)
    return currentWork, errors.Wrap(err, "handlePiece")
}

// getPiece sends requests and receives the piece messages
func (peer *Peer) getPiece(index int) ([]byte, error) {
    // Initialize peer's work
    pieceSize := common.PieceSize(peer.info, index)
    workLeft := pieceSize
    currentWork := make([]byte, pieceSize)

    // start of elapsed time
    for begin := 0; workLeft > 0; {
        // Send max number of requests to peer
        for ; peer.reqsOut < peer.rate; {
            reqSize := common.Min(workLeft, maxReqSize)
            err := peer.sendRequest(index, begin, reqSize)
            if err != nil {
                return nil, errors.Wrap(err, "getPiece")
            }
            begin += reqSize
        }

        // Receive data from the peer
        msg, err := peer.getMessage()
        if err != nil {
            return nil, errors.Wrap(err, "getPiece")
        }
        if _, err = peer.handleMessage(msg, currentWork); err != nil {  // Handle message
            return nil, errors.Wrap(err, "getPiece")
        }
    }

    // end of elapsed time
    return currentWork, nil
}

// downloadPiece starts a routine to download a piece from a peer
func (peer *Peer) downloadPiece(index int) ([]byte, error) {
    // TODO Send interested and receive unchoke
    msg := message.Interested()
    err := peer.Conn.Write(msg.Encode())
    if err != nil {
        return nil, errors.Wrap(err, "downloadPiece")
    }

    piece, err := peer.getPiece(index)
    if err != nil {
        return nil, errors.Wrap(err, "downloadPiece")
    }

    // TODO Send not interested

    // Verify the piece's hash
    if !write.VerifyPiece(peer.info, index, piece) {
        return nil, errors.Wrap(ErrPieceHash, "downloadPiece")
    }
    return piece, nil
}
