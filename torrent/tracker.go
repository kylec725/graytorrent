package torrent

import (
    "strconv"
    "time"
    "math/rand"
    "net/url"
    "net/http"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/pkg/errors"
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
    Complete int
    Incomplete int
    httpClient *http.Client
}

func newTracker(announce string) Tracker {
    return Tracker{
        Announce: announce,
        Working: false,
        Interval: 180,
        Complete: 0,
        Incomplete: 0,
        httpClient: &http.Client{ Timeout: 20 * time.Second },
    }
}

func getTrackers(meta metainfo.BencodeMeta) ([]Tracker, error) {
    // If announce-list is empty, use announce only
    if len(meta.AnnounceList) == 0 {
        // Check if no announce strings exist
        if meta.Announce == "" {
            return nil, errors.Wrap(ErrNoAnnounce, "getTrackers")
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

func (tr Tracker) buildURL(infoHash [20]byte, peerID [20]byte, port uint16, left int, event string) (string, error) {
    base, err := url.Parse(tr.Announce)
    if err != nil {
        return "", errors.Wrap(err, "buildURL")
    }

    params := url.Values{
        "info_hash": []string{string(infoHash[:])},
        "peer_id": []string{string(peerID[:])},
        "port": []string{strconv.Itoa(int(port))},
        "uploaded": []string{"0"},
        "downloaded": []string{"0"},
        "left": []string{strconv.Itoa(left)},
        "compact": []string{"1"},
        "event": []string{event},
    }

    if event == "" {
        delete(params, "event")
    }

    base.RawQuery = params.Encode()
    return base.String(), nil
}
