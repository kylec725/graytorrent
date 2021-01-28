package tracker

import (
    "time"
    "math/rand"
    "net/http"
    "context"

    "github.com/kylec725/graytorrent/common"
    "github.com/kylec725/graytorrent/metainfo"
    "github.com/kylec725/graytorrent/peer"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

// Errors
var (
    ErrNoAnnounce = errors.New("Did not find any annouce urls")
)

// Tracker stores information about a torrent tracker
type Tracker struct {
    Announce string
    Working bool
    Interval int

    txID uint32  // Used by UDP trackers
    cnID uint64

    httpClient *http.Client
}

func newTracker(announce string) Tracker {
    return Tracker{
        Announce: announce,
        Working: false,
        Interval: 2,

        // info: info,
        httpClient: &http.Client{ Timeout: 20 * time.Second },
        // port: port,
        // shutdown: make(chan int),
    }
}

// GetTrackers parses metainfo to retrieve a list of trackers
func GetTrackers(meta metainfo.BencodeMeta, port uint16) ([]Tracker, error) {
    // If announce-list is empty, use announce only
    if len(meta.AnnounceList) == 0 {
        // Check if no announce strings exist
        if meta.Announce == "" {
            return nil, errors.Wrap(ErrNoAnnounce, "GetTrackers")
        }

        trackers := make([]Tracker, 1)
        trackers[0] = newTracker(meta.Announce)
        return trackers, nil
    }

    // Make list of multiple trackers
    var trackers []Tracker
    var numAnnounce int
    // Add each announce in announce-list as a tracker
    for _, group := range meta.AnnounceList {
        for _, announce := range group {
            trackers = append(trackers, newTracker(announce))
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

// Run starts a tracker and gets peers for a torrent
func (tr *Tracker) Run(ctx context.Context, peers chan peer.Peer, complete chan bool) {
    info := common.Info(ctx)
    port := common.Port(ctx)
    trackerLog := log.WithField("tracker", tr.Announce)
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
        currInfo := common.Info(ctx)  // Need to get this because original info is a copy
        if tr.Working {  // Send stopped message if necessary
            uploaded := 0
            downloaded := info.Left - currInfo.Left
            if err = tr.sendStopped(currInfo, port, uploaded, downloaded, currInfo.Left); err != nil {
                trackerLog.WithField("error", err.Error()).Debug("Error while sending stopped message")
            }
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
        case <-time.After(time.Duration(tr.Interval) * time.Second):  // Send announce at intervals
            currInfo := common.Info(ctx)
            uploaded := 0
            downloaded := info.Left - currInfo.Left
            if err = tr.sendAnnounce(currInfo, port, uploaded, downloaded, currInfo.Left); err != nil {
                if tr.Working {  // Reset interval if tracker just stopped working
                    tr.Interval = 2
                } else {  // Double interval when tracker was not previously working
                    tr.Interval *= 2
                }
                tr.Working = false
                trackerLog.WithField("error", err.Error()).Debug("Error while sending announce message")
            } else {
                tr.Working = false
            }
        case _, ok := <-complete:
            if !ok && tr.Working {
                currInfo := common.Info(ctx)
                uploaded := 0
                downloaded := info.Left - currInfo.Left
                if err = tr.sendCompleted(currInfo, port, uploaded, downloaded, currInfo.Left); err != nil {
                    tr.Working = false
                    trackerLog.WithField("error", err.Error()).Debug("Error while sending completed message")
                } else {
                    tr.Working = true
                }
            }
        }
    }
}
