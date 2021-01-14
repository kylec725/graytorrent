package torrent

import (
    "strconv"
    "errors"
    "time"
    "math/rand"
    "net/http"
    "net/url"

    "github.com/kylec725/graytorrent/metainfo"
    "github.com/kylec725/graytorrent/peer"
    bencode "github.com/jackpal/bencode-go"
)

// Tracker stores information about a torrent tracker
type Tracker struct {
    Announce string
    Working bool
    Interval int
}

type bencodeTrackerResp struct {
    Interval int `bencode:"interval"`
    Peers string `bencode:"peers"`
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
    base.RawQuery = params.Encode()

    return base.String(), nil
}

func (tr *Tracker) getPeers(req string) ([]peer.Peer, error) {
    const getRetry = 5
    resp, err := http.Get(req)
    // Resend the GET request several times until we receive a response
    for i := 0; err != nil; resp, err = http.Get(req) {
        if err, ok := err.(*url.Error); !ok {
            return nil, err
        } else if i++; i > getRetry {
            return nil, errors.New("Sent GET requests " + strconv.Itoa(getRetry) + " times with no response")
        }
    }

    if resp.StatusCode != 200 {
        return nil, errors.New("Did not get status code 200: Got status code " + strconv.Itoa(resp.StatusCode))
    }

    // Unmarshal tracker response to get new interval and list of peers
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return nil, err
    }

    tr.Interval = trResp.Interval
    peersBytes := []byte(trResp.Peers)

    return peer.Unmarshal(peersBytes)
}
