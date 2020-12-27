package torrent

import (
    // "fmt"
    "strconv"
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

    trackers := make([]Tracker, 1 + len(meta.AnnounceList))
    trackers[0] = Tracker{meta.Announce, false, 120, to.ID}

    return trackers, nil
}
