package peer

import (
    "encoding/binary"
)

type messageID uint8

const (
    msgChoke         messageID = 0
    msgUnchoke       messageID = 1
    msgInterested    messageID = 2
    msgNotInterested messageID = 3
    masgHave         messageID = 4
    msgBitfield      messageID = 5
    msgRequest       messageID = 6
    msgPiece         messageID = 7
    msgCancel        messageID = 8
    msgPort          messageID = 9
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
