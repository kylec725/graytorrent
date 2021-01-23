package peer

import (
    "encoding/binary"
    "math"
    "fmt"

    "github.com/pkg/errors"
)

type messageID uint8

const (
    msgChoke         messageID = 0
    msgUnchoke       messageID = 1
    msgInterested    messageID = 2
    msgNotInterested messageID = 3
    msgHave          messageID = 4
    msgBitfield      messageID = 5
    msgRequest       messageID = 6
    msgPiece         messageID = 7
    msgCancel        messageID = 8
    msgPort          messageID = 9
)

// Errors
var (
    ErrBitfield = errors.New("Malformed bitfield received")
    ErrSend = errors.New("Unexpected number of bytes sent")
    // ErrRcv = errors.New("Unexpected number of bytes received")
)

// message stores the message type id and payload
type message struct {
    id messageID
    payload []byte
}

func (msg *message) serialize() []byte {
    if msg == nil {  // nil is a keep-alive message
        return make([]byte, 4)
    }
    length := uint32(1 + len(msg.payload))
    serial := make([]byte, length)
    binary.BigEndian.PutUint32(serial[0:4], length)
    serial[4] = byte(msg.id)
    copy(serial[5:], msg.payload)
    return serial
}

func (peer *Peer) rcvMsg() (message, error) {
    buf := make([]byte, 1)
    if _, err := peer.Conn.Read(buf); err != nil {
        return message{}, errors.Wrap(err, "rcvMsg")
    }
    msgLen := buf[0]

    buf = make([]byte, msgLen)
    if _, err := peer.Conn.Read(buf); err != nil {
        return message{}, errors.Wrap(err, "rcvMsg")
    }

    msg := message{id: messageID(buf[0]), payload: buf[1:]}
    return msg, nil
}

func (peer *Peer) sendRequest(index, begin, length int) error {
    payload := make([]byte, 12)
    binary.BigEndian.PutUint32(payload[0:4], uint32(index))
    binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
    binary.BigEndian.PutUint32(payload[8:12], uint32(length))
    msg := message{id: msgRequest, payload: payload}

    bytesSent, err := peer.Conn.Write(msg.serialize())
    if err != nil {
        return errors.Wrap(err, "sendRequest")
    } else if bytesSent != 13 {
        return errors.Wrap(ErrSend, "sendRequest")
    }
    return nil
}

func (peer *Peer) handleMsg(msg *message) error {
    if msg == nil {
        // reset keep-alive
    }
    switch msg.id {
    case msgChoke:
        peer.peerChoking = true
    case msgUnchoke:
        peer.peerChoking = false
    case msgInterested:
        peer.peerInterested = true
    case msgNotInterested:
        peer.peerInterested = false
    case msgHave:
        index := binary.BigEndian.Uint32(msg.payload)
        peer.Bitfield.Set(int(index))
    case msgBitfield:
        expected := int(math.Ceil(float64(peer.info.TotalPieces) / 8))
        if expected != len(msg.payload) {
            return errors.Wrap(ErrBitfield, "handleMsg")
        }
        peer.Bitfield = msg.payload
    case msgRequest:
        return errors.New("Not yet implemented")
    case msgPiece:
        return errors.New("Not yet implemented")
    case msgCancel:
        fmt.Println("msgPort not yet implemented")
    case msgPort:
        fmt.Println("msgPort not yet implemented")
    }
    return nil
}

func (peer *Peer) handleRequest(msg *message) error {
    return nil
}

func (peer *Peer) handlePiece(msg *message) error {
    // index := binary.BigEndian.Uint32(msg.payload[0:4])
    // begin := binary.BigEndian.Uint32(msg.payload[4:8])
    // block := msg.payload[8:]

    return nil
}
