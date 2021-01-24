package message

import (
    "encoding/binary"

    "github.com/kylec725/graytorrent/bitfield"
)

type messageID uint8

// Message types
const (
    MsgChoke         messageID = 0
    MsgUnchoke       messageID = 1
    MsgInterested    messageID = 2
    MsgNotInterested messageID = 3
    MsgHave          messageID = 4
    MsgBitfield      messageID = 5
    MsgRequest       messageID = 6
    MsgPiece         messageID = 7
    MsgCancel        messageID = 8
    MsgPort          messageID = 9
)

// Message stores the message type id and payload
type Message struct {
    ID messageID
    Payload []byte
}

// Encode serializes a message for sending over a wire
func (msg *Message) Encode() []byte {
    if msg == nil {  // nil is a keep-alive message
        return make([]byte, 4)
    }
    length := uint32(1 + len(msg.Payload))
    serial := make([]byte, length)
    binary.BigEndian.PutUint32(serial[0:4], length)
    serial[4] = byte(msg.ID)
    copy(serial[5:], msg.Payload)
    return serial
}

// Decode decodes a byte slice into a message
func Decode(data []byte) *Message {
    if data == nil {
        return nil
    }
    id := messageID(data[0])
    payload := data[1:]
    return &Message{ID: id, Payload: payload}
}

// Choke returns a choke message
func Choke() Message {
    return Message{ID: MsgChoke}
}

// Unchoke returns an unchoke message
func Unchoke() Message {
    return Message{ID: MsgUnchoke}
}

// Interested returns an interested message
func Interested() Message {
    return Message{ID: MsgInterested}
}

// NotInterested returns a not interested message
func NotInterested() Message {
    return Message{ID: MsgNotInterested}
}

// Have returns a have message
func Have(index uint32) Message {
    payload := make([]byte, 4)
    binary.BigEndian.PutUint32(payload, index)
    return Message{ID: MsgNotInterested, Payload: payload}
}

// Bitfield returns a bitfield message
func Bitfield(bf bitfield.Bitfield) Message {
    return Message{ID: MsgBitfield, Payload: bf}
}

// Request returns a request message for a piece
func Request(index, begin, length uint32) Message {
    payload := make([]byte, 12)
    binary.BigEndian.PutUint32(payload[0:4], index)
    binary.BigEndian.PutUint32(payload[4:8], begin)
    binary.BigEndian.PutUint32(payload[8:12], length)
    return Message{ID: MsgRequest, Payload: payload}
}

// Piece returns a piece message containing a block
func Piece(index, begin uint32, block []byte) Message {
    payload := make([]byte, 4 + len(block))
    binary.BigEndian.PutUint32(payload, index)
    copy(payload[4:], block)
    return Message{ID: MsgPiece, Payload: payload}
}
