package tracker

import (
    "math/rand"
    "time"
    "encoding/binary"
    "fmt"

    "github.com/kylec725/graytorrent/connect"
    "github.com/pkg/errors"
)

const udpTimeout = 10 * time.Second

// Errors
var (
    ErrUDP = errors.New("Unexpected number of packets send through UDP")
)

func (tr *Tracker) udpConnect(conn connect.Conn, transactionID uint32) error {
    packet := make([]byte, 16)
    connectionID := uint64(0x41727101980)
    action := uint32(0)
    binary.BigEndian.PutUint64(packet[0:8], connectionID)
    binary.BigEndian.PutUint32(packet[8:12], action)
    binary.BigEndian.PutUint32(packet[12:16], transactionID)
    err := conn.Write(packet)
    if err != nil {
        return errors.Wrap(err, "udpConnect")
    }

    // Response
    packet = make([]byte, 16)
    err = conn.ReadFull(packet)
    if err != nil {
        return errors.Wrap(err, "udpConnect")
    }
    action = binary.BigEndian.Uint32(packet[0:4])
    

    return nil
}

func (tr *Tracker) udpAnnounce(conn connect.Conn) error {
    rand.Seed(time.Now().UnixNano())
    packet := make([]byte, 72)


    err := conn.Write(packet)
    if err != nil{
        return errors.Wrap(err, "udpConnect")
    }

    return nil
}


func (tr *Tracker) udpStarted() ([]byte, error) {
    rand.Seed(time.Now().UnixNano())
    transactionID := rand.Int31()
    fmt.Println("transactionID", transactionID)

    return nil, nil
}

func (tr *Tracker) udpStopped() error {

    return nil
}
