package torrent

import (
    "strconv"
    "errors"
    "time"
    "math/rand"
    "net/url"

    "github.com/kylec725/graytorrent/metainfo"
)

// Tracker stores information about a torrent tracker
type Tracker struct {
    Announce string
    Working bool
    Interval int
}

type bencodeTrackerResp struct {
    Interval int `bencode:"interval"`
    Peers string `bencode:"peers,omitempty"`
    Failure string `bencode:"failure reason,omitempty"`
    Complete int `bencode:"complete"`
    Incomplete int `bencode:"incomplete"`
}

func (tr Tracker) String() string {
    var result string
    result += "Announce: " + tr.Announce + "\n"
    result += "Working: " + strconv.FormatBool(tr.Working) + "\n"
    result += "Interval: " + strconv.Itoa(tr.Interval) + "\n"

    return result
}

func getTrackers(meta metainfo.BencodeMeta) ([]Tracker, error) {
    // Get meta data of torrent file first
    numAnnounce := 0
    for _, group := range meta.AnnounceList {
        numAnnounce += len(group)
    }

    // If announce-list empty, use announce only
    if numAnnounce == 0 {
        // Check if no announce strings exist
        if meta.Announce == "" {
            return nil, errors.New("Did not get any announce urls")
        }

        trackers := make([]Tracker, 1)
        trackers[0] = Tracker{meta.Announce, false, 180}
        return trackers, nil
    }
    
    // Make list of multiple trackers
    trackers := make([]Tracker, numAnnounce)
    // Loop through announce-list
    i := 0
    for _, group := range meta.AnnounceList {
        for _, announce := range group {
            trackers[i] = Tracker{announce, false, 180}
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

func (tr Tracker) buildURL(infoHash [20]byte, peerID [20]byte, port uint16, left int, event string) (string, error) {
    base, err := url.Parse(tr.Announce)
    if err != nil {
        return "", err
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
