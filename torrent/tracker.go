package torrent

import (
    // "fmt"
    "strconv"
    "errors"
    "time"
    "math/rand"
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

func (to Torrent) getTrackers() ([]Tracker, error) {
    // Get meta data of torrent file first
    meta, err := to.getMeta()
    if err != nil {
        return nil, err
    }

    numAnnounce := 0
    for _, group := range meta.AnnounceList {
        numAnnounce += len(group)
    }

    // If announce-list empty, use announce only
    if numAnnounce == 0 {
        // Check if announce list is empty
        if meta.Announce == "" {
            return nil, errors.New("Did not get any announce urls")
        }

        trackers := make([]Tracker, 1)
        trackers[0] = Tracker{meta.Announce, false, 120, to.ID}
        return trackers, nil
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

    return trackers, nil
}

func (tr Tracker) buildURL() (string, error) {

    return "", nil
}
