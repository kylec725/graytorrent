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
    ErrMessage = errors.New("Malformed message received")
)

// message stores the message type id and payload
type message struct {
    id messageID
    payload []byte
}

func (msg *message) encode() []byte {
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

func decode(data []byte) message {
    id := messageID(data[0])
    payload := data[1:]
    return message{id, payload}
}

// TODO replace with generic reading from connection
// CURRENTLY UNUSED (REMOVE?)
func (peer *Peer) rcvMessage() (*message, error) {
    buf := make([]byte, 4)
    if err := peer.Conn.Read(buf); err != nil {
        return nil, errors.Wrap(err, "rcvMessage")
    }
    msgLen := binary.BigEndian.Uint32(buf)
    if msgLen == 0 {  // Keep-alive message
        return nil, nil
    }

    buf = make([]byte, msgLen)
    if err := peer.Conn.Read(buf); err != nil {
        return nil, errors.Wrap(err, "rcvMessage")
    }

    msg := decode(buf)
    return &msg, nil
}

func (peer *Peer) sendRequest(index, begin, length int) error {
    payload := make([]byte, 12)
    binary.BigEndian.PutUint32(payload[0:4], uint32(index))
    binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
    binary.BigEndian.PutUint32(payload[8:12], uint32(length))
    msg := message{id: msgRequest, payload: payload}

    err := peer.Conn.Write(msg.encode())
    return errors.Wrap(err, "sendRequest")
}

func (peer *Peer) handleMessage(msg *message) error {
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
            return errors.Wrap(ErrBitfield, "handleMessage")
        }
        peer.Bitfield = msg.payload
    case msgRequest:
        if !peer.amChoking {

        }
        return errors.New("Not yet implemented")
    case msgPiece:
        // discard because we did not explicitly request it
        return errors.New("Not yet implemented")
    case msgCancel:
        fmt.Println("msgPort not yet implemented")
    case msgPort:
        fmt.Println("msgPort not yet implemented")
    }
    return nil
}

// TODO
func (peer *Peer) handleRequest(msg *message) error {
    return nil
}

// TODO
func (peer *Peer) handlePiece(msg *message) error {
    // index := binary.BigEndian.Uint32(msg.payload[0:4])
    // begin := binary.BigEndian.Uint32(msg.payload[4:8])
    // block := msg.payload[8:]

    return nil
}
