package tracker

import (
	"context"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/metainfo"
	"github.com/kylec725/graytorrent/internal/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const numWant = 30 // Max peers to receive in tracker response

// TODO: test that tracker messages have correct numbers

// Errors
var (
	ErrNoAnnounce = errors.New("Did not find any annouce urls")
)

// Tracker stores information about a torrent tracker
type Tracker struct {
	Announce string `json:"Announce"`
	Working  bool   `json:"-"`
	Interval int    `json:"-"`

	conn *net.UDPConn `json:"-"` // Used by UDP trackers
	txID uint32       `json:"-"`
	cnID uint64       `json:"-"`

	httpClient *http.Client `json:"-"`
}

// NOTE: consider structuring trackers as an interface and separate http vs udp trackers

// New returns a new tracker
func New(announce string) *Tracker {
	return &Tracker{
		Announce: announce,
		Working:  false,
		Interval: 2,

		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

// TODO: change to use tiers of trackers

// GetTrackers parses metainfo to retrieve a list of trackers
func GetTrackers(m metainfo.Metainfo) ([]*Tracker, error) {
	// If announce-list is empty, use announce only
	if len(m.AnnounceList) == 0 {
		// Check if no announce strings exist
		if m.Announce == "" {
			return nil, errors.Wrap(ErrNoAnnounce, "GetTrackers")
		}

		trackers := make([]*Tracker, 1)
		trackers[0] = New(m.Announce)
		return trackers, nil
	}

	// Make list of multiple trackers
	var trackers []*Tracker
	var numAnnounce int
	// Add each announce in announce-list as a tracker
	for _, group := range m.AnnounceList {
		for _, announce := range group {
			if len(announce) < 4 {
				continue
			} else if announce[:4] != "http" && announce[:3] != "udp" {
				continue
			}
			trackers = append(trackers, New(announce))
			numAnnounce++
		}
	}

	// Shuffle list before returning
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(numAnnounce, func(x, y int) {
		trackers[x], trackers[y] = trackers[y], trackers[x]
	})

	return trackers, nil
}

func (tr *Tracker) sendStarted(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]peer.Peer, error) {
	if tr.Announce[:4] == "http" {
		peerList, err := tr.httpStarted(info, port, uploaded, downloaded, left)
		return peerList, errors.Wrap(err, "sendStarted")
	}
	if err := tr.udpConnect(); err != nil {
		return nil, errors.Wrap(err, "sendStarted")
	}
	peerList, err := tr.udpStarted(info, port, uploaded, downloaded, left)
	return peerList, errors.Wrap(err, "sendStarted")
}

func (tr *Tracker) sendStopped(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) error {
	if tr.Announce[:4] == "http" {
		err := tr.httpStopped(info, port, uploaded, downloaded, left)
		return errors.Wrap(err, "sendStopped")
	}
	if tr.conn == nil {
		if err := tr.udpConnect(); err != nil {
			return errors.Wrap(err, "sendStopped")
		}
	}
	err := tr.udpStopped(info, port, uploaded, downloaded, left)
	return errors.Wrap(err, "sendStopped")
}

func (tr *Tracker) sendCompleted(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) error {
	if tr.Announce[:4] == "http" {
		err := tr.httpCompleted(info, port, uploaded, downloaded, left)
		return errors.Wrap(err, "sendCompleted")
	}
	if tr.conn == nil {
		if err := tr.udpConnect(); err != nil {
			return errors.Wrap(err, "sendCompleted")
		}
	}
	err := tr.udpCompleted(info, port, uploaded, downloaded, left)
	return errors.Wrap(err, "sendCompleted")
}

func (tr *Tracker) sendAnnounce(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]peer.Peer, error) {
	if tr.Announce[:4] == "http" {
		peerList, err := tr.httpAnnounce(info, port, uploaded, downloaded, left)
		return peerList, errors.Wrap(err, "sendAnnounce")
	}
	if tr.conn == nil {
		if err := tr.udpConnect(); err != nil {
			return nil, errors.Wrap(err, "sendAnnounce")
		}
	}
	peerList, err := tr.udpAnnounce(info, port, uploaded, downloaded, left)
	return peerList, errors.Wrap(err, "sendAnnounce")
}

// Run starts a tracker and gets peers for a torrent
func (tr *Tracker) Run(ctx context.Context, info *common.TorrentInfo, peers chan peer.Peer, complete chan bool) {
	startLeft := info.Left
	port := common.Port(ctx)
	trackerLog := log.WithField("tracker", tr.Announce)
	tr.Interval = 2 // Initialize values that can't be saved
	tr.httpClient = &http.Client{Timeout: 20 * time.Second}

	peerList, err := tr.sendStarted(info, port, 0, 0, info.Left)
	if err != nil {
		tr.Working = false
		trackerLog.WithField("error", err.Error()).Debug("Error while sending started message")
	} else {
		tr.Working = true
		trackerLog.WithField("amount", len(peerList)).Debug("Received list of peers")
	}

	// Cleanup
	defer func() {
		if tr.Working { // Send stopped message if necessary
			uploaded := 0
			downloaded := info.Left - startLeft
			if err = tr.sendStopped(info, port, uploaded, downloaded, info.Left); err != nil {
				trackerLog.WithField("error", err.Error()).Debug("Error while sending stopped message")
			}
		}
		if tr.conn != nil { // Close connections for UDP trackers
			tr.conn.Close()
		}
	}()

	// Send peers through channel
	for i := range peerList {
		peers <- peerList[i]
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(tr.Interval) * time.Second): // Send announce at intervals
			uploaded := 0
			downloaded := info.Left - startLeft
			if peerList, err = tr.sendAnnounce(info, port, uploaded, downloaded, info.Left); err != nil {
				if tr.Working { // Reset interval if tracker just stopped working
					tr.Interval = 2
				} else { // Double interval if tracker was already not working
					tr.Interval *= 2
				}
				tr.Working = false
			} else {
				trackerLog.WithField("amount", len(peerList)).Debug("Received list of peers")
				// Send peers through channel
				for i := range peerList {
					peers <- peerList[i]
				}
				tr.Working = true
			}
		case _, ok := <-complete: // WARNING: if we don't return here, this case will loop
			if !ok && tr.Working {
				uploaded := 0
				downloaded := info.Left - startLeft
				if err = tr.sendCompleted(info, port, uploaded, downloaded, info.Left); err != nil {
					tr.Working = false
					trackerLog.WithField("error", err.Error()).Debug("Error while sending completed message")
				} else {
					tr.Working = true
				}
				return // TODO: find way to have tracker continue sending messages when complete without loop
			}
		}
	}
}
