package torrent

import (
    // "fmt"
    "strconv"
    "errors"
    "time"
    "math/rand"

    "github.com/kylec725/graytorrent/metainfo"
)

// Tracker stores information about a torrent tracker
type Tracker struct {
    Announce string
    Working bool
    Interval int
    ID [20]byte
}

func (tr Tracker) String() string {
    var result string
    result += "Announce: " + tr.Announce + "\n"
    result += "Working: " + strconv.FormatBool(tr.Working) + "\n"
    result += "Interval: " + strconv.Itoa(tr.Interval) + "\n"

    return result
}

func (to *Torrent) getTrackers(meta metainfo.BencodeMeta) error {
    // Get meta data of torrent file first
    numAnnounce := 0
    for _, group := range meta.AnnounceList {
        numAnnounce += len(group)
    }

    // If announce-list empty, use announce only
    if numAnnounce == 0 {
        // Check if announce list is empty
        if meta.Announce == "" {
            return errors.New("Did not get any announce urls")
        }

        trackers := make([]Tracker, 1)
        trackers[0] = Tracker{meta.Announce, false, 120, to.ID}
        return nil
    }
    
    // Make list of multiple trackers
    trackers := make([]Tracker, numAnnounce)
    // Loop through announce-list
    i := 0
    for _, group := range meta.AnnounceList {
        for _, announce := range group {
            trackers[i] = Tracker{announce, false, 120, to.ID}
            i++
        }
    }

    // Shuffle list before returning
    rand.Seed(time.Now().UnixNano())
    rand.Shuffle(numAnnounce, func(x, y int) {
        trackers[x], trackers[y] = trackers[y], trackers[x]
    })

    // return trackers, nil
    to.Trackers = trackers
    return nil
}

func (tr Tracker) buildURL() string {

    return ""
}
