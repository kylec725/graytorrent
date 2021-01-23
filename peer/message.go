package peer

import (
    "encoding/binary"
    "math"
    "fmt"

    "github.com/kylec725/graytorrent/bitfield"
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

// Message stores the message type id and payload
type Message struct {
    ID messageID
    Payload []byte
}

func (msg *Message) encode() []byte {
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

func decode(data []byte) Message {
    id := messageID(data[0])
    payload := data[1:]
    return Message{id, payload}
}

// Choke returns a choke message
func Choke() Message {
    return Message{ID: msgChoke}
}

// Unchoke returns an unchoke message
func Unchoke() Message {
    return Message{ID: msgUnchoke}
}

// Interested returns an interested message
func Interested() Message {
    return Message{ID: msgInterested}
}

// NotInterested returns a not interested message
func NotInterested() Message {
    return Message{ID: msgNotInterested}
}

// Have returns a have message
// TODO
func Have(index int) Message {
    payload := make([]byte, 4)
    binary.BigEndian.PutUint32(payload, uint32(index))
    return Message{ID: msgNotInterested, Payload: payload}
}

// Bitfield returns a bitfield message
// TODO
func Bitfield(bf bitfield.Bitfield) Message {
    return Message{}
}

// Request returns a request message for a piece
// TODO
func Request(index, begin, length int) Message {
    return Message{}
}

// Piece returns a piece message containing a block
// TODO
func Piece(index, begin int, block []byte) Message {
    return Message{}
}

// TODO replace with generic reading from connection
func (peer *Peer) rcvMessage() (*Message, error) {
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
    msg := Message{ID: msgRequest, Payload: payload}

    err := peer.Conn.Write(msg.encode())
    return errors.Wrap(err, "sendRequest")
}

func (peer *Peer) handleMessage(msg *Message) error {
    if msg == nil {
        // reset keep-alive
    }
    switch msg.ID {
    case msgChoke:
        peer.peerChoking = true
    case msgUnchoke:
        peer.peerChoking = false
    case msgInterested:
        peer.peerInterested = true
    case msgNotInterested:
        peer.peerInterested = false
    case msgHave:
        index := binary.BigEndian.Uint32(msg.Payload)
        peer.Bitfield.Set(int(index))
    case msgBitfield:
        expected := int(math.Ceil(float64(peer.info.TotalPieces) / 8))
        if expected != len(msg.Payload) {
            return errors.Wrap(ErrBitfield, "handleMessage")
        }
        peer.Bitfield = msg.Payload
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
func (peer *Peer) handleRequest(msg *Message) error {
    return nil
}

// TODO
func (peer *Peer) handlePiece(msg *Message) error {
    // index := binary.BigEndian.Uint32(msg.payload[0:4])
    // begin := binary.BigEndian.Uint32(msg.payload[4:8])
    // block := msg.payload[8:]

    return nil
}
