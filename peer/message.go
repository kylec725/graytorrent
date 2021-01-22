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
    return message{}, nil
}

func (peer *Peer) handleMsg(msg message) error {
    switch msg.id {
    case msgChoke:
        peer.PeerChoking = true
    case msgUnchoke:
        peer.PeerChoking = false
    case msgInterested:
        peer.PeerInterested = true
    case msgNotInterested:
        peer.PeerInterested = false
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
