package torrent

import (
    "fmt"
    "strconv"
    "errors"
    "time"
    "math/rand"
    // "net"
    "net/http"
    "net/url"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/kylec725/graytorrent/peer"
)

// Tracker stores information about a torrent tracker
type Tracker struct {
    Announce string
    Working bool
    Interval int
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
            return []Tracker{}, errors.New("Did not get any announce urls")
        }

        trackers := make([]Tracker, 1)
        trackers[0] = Tracker{meta.Announce, false, 120}
        return trackers, nil
    }
    
    // Make list of multiple trackers
    trackers := make([]Tracker, numAnnounce)
    // Loop through announce-list
    i := 0
    for _, group := range meta.AnnounceList {
        for _, announce := range group {
            trackers[i] = Tracker{announce, false, 120}
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
    base.RawQuery = params.Encode()

    return base.String(), nil
}

func (tr Tracker) getPeers(req string) ([]peer.Peer, error) {
    resp, err := http.Get(req)
    // Resend the GET request until we receive a response
    for ; err != nil; resp, err = http.Get(req) {
        if err, ok := err.(*url.Error); !ok {
            return []peer.Peer{}, err
        }
    }

    // TODO: parse the response
    fmt.Println("resp:", resp)
    return []peer.Peer{}, nil
}
