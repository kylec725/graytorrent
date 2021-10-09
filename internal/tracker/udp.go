package tracker

import (
	"encoding/binary"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/peer"
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
	ErrEvent        = errors.New("Tried to create packet with invalid event")
)

func (tr *Tracker) udpAddr() string {
	splitAddr := strings.Split(tr.Announce, "/")
	if len(splitAddr) < 3 {
		return ""
	}
	return splitAddr[2]
}

// udpConnect initializes a UDP exchange with a tracker
func (tr *Tracker) udpConnect() error {
	rand.Seed(time.Now().UnixNano())
	tr.txID = rand.Uint32()

	addr := tr.udpAddr()
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	}
	tr.conn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	}

	// Request
	req := make([]byte, 16)
	connectionID := uint64(0x41727101980)
	action := uint32(0)
	binary.BigEndian.PutUint64(req[0:8], connectionID) // Protocol ID
	binary.BigEndian.PutUint32(req[8:12], action)      // Action: Connection
	binary.BigEndian.PutUint32(req[12:16], tr.txID)    // Transaction ID
	_, err = tr.conn.Write(req)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	}

	// Response
	resp := make([]byte, 16)
	bytesRead, err := tr.conn.Read(resp)
	if err != nil {
		return errors.Wrap(err, "udpConnect")
	} else if bytesRead < 16 {
		return errors.Wrap(ErrSize, "udpConnect")
	}
	action = binary.BigEndian.Uint32(resp[0:4])   // Action
	txID := binary.BigEndian.Uint32(resp[4:8])    // Transaction ID
	tr.cnID = binary.BigEndian.Uint64(resp[8:16]) // Connection ID

	// Verify response
	if action == 3 {
		errorString := string(resp[8:])
		log.WithFields(log.Fields{"tracker": tr.Announce, "message": errorString}).Debug("Got error message from tracker")
		return errors.Wrap(ErrTrackerError, "udpConnect")
	} else if action != 0 {
		return errors.Wrap(ErrAction, "udpConnect")
	} else if txID != tr.txID {
		return errors.Wrap(ErrTransaction, "udpConnect")
	}
	return nil
}

// buildPacket creates an announce packet for a corresponding event
func (tr *Tracker) buildPacket(event string, info common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]byte, error) {
	var eventCode uint32
	switch event {
	case "announce":
		eventCode = 0
	case "completed":
		eventCode = 1
	case "started":
		eventCode = 2
	case "stopped":
		eventCode = 3
	default:
		return nil, errors.Wrap(ErrEvent, "buildPacket")
	}
	rand.Seed(time.Now().UnixNano())
	key := rand.Uint32()

	packet := make([]byte, 100)
	binary.BigEndian.PutUint64(packet[0:8], tr.cnID)              // Connection ID
	binary.BigEndian.PutUint32(packet[8:12], 1)                   // Action: Announce
	binary.BigEndian.PutUint32(packet[12:16], tr.txID)            // Transaction ID
	copy(packet[16:36], info.InfoHash[:])                         // Info Hash
	copy(packet[36:56], info.PeerID[:])                           // Peer ID
	binary.BigEndian.PutUint64(packet[56:64], uint64(downloaded)) // Downloaded
	binary.BigEndian.PutUint64(packet[64:72], uint64(left))       // Left
	binary.BigEndian.PutUint64(packet[72:80], uint64(uploaded))   // Uploaded
	binary.BigEndian.PutUint32(packet[80:84], eventCode)          // Event
	binary.BigEndian.PutUint32(packet[84:88], uint32(0))          // IP Address
	binary.BigEndian.PutUint32(packet[88:92], key)                // Key
	binary.BigEndian.PutUint32(packet[92:96], uint32(numWant))    // Max peers we want
	binary.BigEndian.PutUint16(packet[96:98], port)               // Port
	binary.BigEndian.PutUint16(packet[98:100], uint16(0))         // Extensions
	return packet, nil
}

func (tr *Tracker) udpStarted(info common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]peer.Peer, error) {
	// Request
	req, err := tr.buildPacket("started", info, port, uploaded, downloaded, left)
	if err != nil {
		return nil, errors.Wrap(err, "udpStarted")
	}

	_, err = tr.conn.Write(req)
	if err != nil {
		return nil, errors.Wrap(err, "udpStarted")
	}

	// Response
	resp := make([]byte, 20+6*numWant)
	bytesRead, err := tr.conn.Read(resp)
	if err != nil {
		return nil, errors.Wrap(err, "udpStarted")
	} else if bytesRead < 20 {
		return nil, errors.Wrap(ErrSize, "udpStarted")
	}
	action := binary.BigEndian.Uint32(resp[0:4])    // Action
	txID := binary.BigEndian.Uint32(resp[4:8])      // Transaction ID
	interval := binary.BigEndian.Uint32(resp[8:12]) // New tracker interval

	// Verify response
	if action == 3 {
		errorString := string(resp[8:])
		log.WithFields(log.Fields{"tracker": tr.Announce, "message": errorString}).Debug("Got error message from tracker")
		return nil, errors.Wrap(ErrTrackerError, "udpStarted")
	} else if action != 1 {
		return nil, errors.Wrap(ErrAction, "udpStarted")
	} else if txID != tr.txID {
		return nil, errors.Wrap(ErrTransaction, "udpStarted")
	}

	// Update tracker information
	tr.Interval = int(interval)

	// Get peer information
	peersBytes := resp[20:]
	peersList, err := peer.Unmarshal(peersBytes, info)
	return peersList, errors.Wrap(err, "udpStarted")
}

func (tr *Tracker) udpStopped(info common.TorrentInfo, port uint16, uploaded, downloaded, left int) error {
	// Request
	req, err := tr.buildPacket("stopped", info, port, uploaded, downloaded, left)
	if err != nil {
		return errors.Wrap(err, "udpStopped")
	}

	_, err = tr.conn.Write(req)
	if err != nil {
		return errors.Wrap(err, "udpStopped")
	}

	// Response
	resp := make([]byte, 20+6*numWant)
	bytesRead, err := tr.conn.Read(resp)
	if err != nil {
		return errors.Wrap(err, "udpStopped")
	} else if bytesRead < 20 {
		return errors.Wrap(ErrSize, "udpStopped")
	}
	action := binary.BigEndian.Uint32(resp[0:4])    // Action
	txID := binary.BigEndian.Uint32(resp[4:8])      // Transaction ID
	interval := binary.BigEndian.Uint32(resp[8:12]) // New tracker interval

	// Verify response
	if action == 3 {
		errorString := string(resp[8:])
		log.WithFields(log.Fields{"tracker": tr.Announce, "message": errorString}).Debug("Got error message from tracker")
		return errors.Wrap(ErrTrackerError, "udpStopped")
	} else if action != 1 {
		return errors.Wrap(ErrAction, "udpStopped")
	} else if txID != tr.txID {
		return errors.Wrap(ErrTransaction, "udpStopped")
	}

	// Update tracker information
	tr.Interval = int(interval)

	return nil
}

func (tr *Tracker) udpCompleted(info common.TorrentInfo, port uint16, uploaded, downloaded, left int) error {
	// Request
	req, err := tr.buildPacket("completed", info, port, uploaded, downloaded, left)
	if err != nil {
		return errors.Wrap(err, "udpCompleted")
	}

	_, err = tr.conn.Write(req)
	if err != nil {
		return errors.Wrap(err, "udpCompleted")
	}

	// Response
	resp := make([]byte, 20+6*numWant)
	bytesRead, err := tr.conn.Read(resp)
	if err != nil {
		return errors.Wrap(err, "udpCompleted")
	} else if bytesRead < 20 {
		return errors.Wrap(ErrSize, "udpCompleted")
	}
	action := binary.BigEndian.Uint32(resp[0:4])    // Action
	txID := binary.BigEndian.Uint32(resp[4:8])      // Transaction ID
	interval := binary.BigEndian.Uint32(resp[8:12]) // New tracker interval

	// Verify response
	if action == 3 {
		errorString := string(resp[8:])
		log.WithFields(log.Fields{"tracker": tr.Announce, "message": errorString}).Debug("Got error message from tracker")
		return errors.Wrap(ErrTrackerError, "udpCompleted")
	} else if action != 1 {
		return errors.Wrap(ErrAction, "udpCompleted")
	} else if txID != tr.txID {
		return errors.Wrap(ErrTransaction, "udpCompleted")
	}

	// Update tracker information
	tr.Interval = int(interval)

	return nil
}

func (tr *Tracker) udpAnnounce(info common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]peer.Peer, error) {
	// Request
	req, err := tr.buildPacket("announce", info, port, uploaded, downloaded, left)
	if err != nil {
		return nil, errors.Wrap(err, "udpAnnounce")
	}

	_, err = tr.conn.Write(req)
	if err != nil {
		return nil, errors.Wrap(err, "udpAnnounce")
	}

	// Response
	resp := make([]byte, 20+6*numWant)
	bytesRead, err := tr.conn.Read(resp)
	if err != nil {
		return nil, errors.Wrap(err, "udpAnnounce")
	} else if bytesRead < 20 {
		return nil, errors.Wrap(ErrSize, "udpAnnounce")
	}
	action := binary.BigEndian.Uint32(resp[0:4])    // Action
	txID := binary.BigEndian.Uint32(resp[4:8])      // Transaction ID
	interval := binary.BigEndian.Uint32(resp[8:12]) // New tracker interval

	// Verify response
	if action == 3 {
		errorString := string(resp[8:])
		log.WithFields(log.Fields{"tracker": tr.Announce, "message": errorString}).Debug("Got error message from tracker")
		return nil, errors.Wrap(ErrTrackerError, "udpAnnounce")
	} else if action != 1 {
		return nil, errors.Wrap(ErrAction, "udpAnnounce")
	} else if txID != tr.txID {
		return nil, errors.Wrap(ErrTransaction, "udpAnnounce")
	}

	// Update tracker information
	tr.Interval = int(interval)

	// Get peer information
	peersBytes := resp[20:]
	peersList, err := peer.Unmarshal(peersBytes, info)
	return peersList, errors.Wrap(err, "udpStarted")
}
