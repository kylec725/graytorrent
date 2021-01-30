package tracker

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/kylec725/graytorrent/common"
	"github.com/kylec725/graytorrent/connect"
	"github.com/kylec725/graytorrent/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const udpTimeout = 10 * time.Second

// Errors
var (
	ErrSize         = errors.New("Got packet with unexpected size")
	ErrTransaction  = errors.New("Received incorrect transaction ID")
	ErrAction       = errors.New("Got wrong action code from the tracker")
	ErrTrackerError = errors.New("Received an error message from the tracker")
)

func (tr *Tracker) udpConnect() error {
	// Initialize
	rand.Seed(time.Now().UnixNano())
	tr.txID = rand.Uint32()

	addr, err := net.ResolveUDPAddr("udp", tr.Announce)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	}
	tr.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	}

	// Send connect packet
	packet := make([]byte, 16)
	connectionID := uint64(0x41727101980)
	action := uint32(0)
	binary.BigEndian.PutUint64(packet[0:8], connectionID) // Protocol ID
	binary.BigEndian.PutUint32(packet[8:12], action)      // Action: Connection
	binary.BigEndian.PutUint32(packet[12:16], tr.txID)    // Transaction ID
	_, err = tr.conn.Write(packet)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	}

	// Response
	packet = make([]byte, 16)
	bytesRead, err = tr.conn.Read(packet)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	} else if bytesRead < 16 {
		return errors.Wrap(ErrSize, "udpConnect")
	}
	action = binary.BigEndian.Uint32(packet[0:4])   // Action
	txID := binary.BigEndian.Uint32(packet[4:8])    // Transaction ID
	tr.cnID = binary.BigEndian.Uint64(packet[8:16]) // Connection ID

	// Verify response
	if action == 3 {
		errorString := string(packet[8:])
		log.WithFields(log.Fields{"tracker": tr.Announce, "message": errorString}).Debug("Got error message from tracker")
		return errors.Wrap(ErrTrackerError, "udpConnect")
	} else if action != 1 {
		return errors.Wrap(ErrAction, "udpConnect")
	} else if txID != tr.txID {
		return errors.Wrap(ErrTransaction, "udpConnect")
	}
	return nil
}

func (tr *Tracker) udpStarted(info common.TorrentInfo, port uint16, conn connect.Conn) ([]peer.Peer, error) {
	rand.Seed(time.Now().UnixNano())
	action := uint32(1)
	key := rand.Uint32()
	packet := make([]byte, 100)
	binary.BigEndian.PutUint64(packet[0:8], tr.cnID)                              // Connection ID
	binary.BigEndian.PutUint32(packet[8:12], action)                              // Action: Announce
	binary.BigEndian.PutUint32(packet[12:16], tr.txID)                            // Transaction ID
	copy(packet[16:36], info.InfoHash[:])                                         // Info Hash
	copy(packet[36:56], info.PeerID[:])                                           // Peer ID
	binary.BigEndian.PutUint64(packet[56:64], uint64(info.TotalLength-info.Left)) // Downloaded
	binary.BigEndian.PutUint64(packet[64:72], uint64(info.Left))                  // Left
	binary.BigEndian.PutUint64(packet[72:80], uint64(0))                          // Uploaded
	binary.BigEndian.PutUint32(packet[80:84], uint32(2))                          // Event
	binary.BigEndian.PutUint32(packet[84:88], uint32(0))                          // IP Address
	binary.BigEndian.PutUint32(packet[88:92], key)                                // Key
	binary.BigEndian.PutUint32(packet[92:96], uint32(30))                         // Max peers we want
	binary.BigEndian.PutUint16(packet[96:98], port)                               // Port
	binary.BigEndian.PutUint16(packet[98:100], uint16(0))                         // Extensions

	_, err := conn.Write(packet)
	if err != nil {
		return nil, errors.Wrap(err, "udpStarted")
	}

	// Response
	packet = make([]byte, 16)
	bytesRead, err = tr.conn.Read(packet)
	if err != nil {
		return errors.Wrap(err, "udpStarted")
	} else if bytesRead < 20 {
		return errors.Wrap(ErrSize, "udpStarted")
	}

	return nil, nil
}

func (tr *Tracker) udpStopped() error {

	return nil
}
