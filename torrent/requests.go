package torrent

import (
    "net/http"
    "net/url"

    "github.com/kylec725/graytorrent/peer"
    "github.com/pkg/errors"
    bencode "github.com/jackpal/bencode-go"
)

const reqRetry = 5 // Number of times to retry sending GET requests if first attempt fails

// Errors
var (
    ErrBadStatusCode = errors.New("Expected status code 200")
)

type bencodeTrackerResp struct {
    Interval int `bencode:"interval"`
    Peers string `bencode:"peers"`
    Failure string `bencode:"failure reason"`
    Complete int `bencode:"complete"`
    Incomplete int `bencode:"incomplete"`
}

func (tr *Tracker) sendStarted(infoHash [20]byte, peerID [20]byte, port uint16, left int) ([]peer.Peer, error) {
    req, err := tr.buildURL(infoHash, peerID, port, left, "started")
    if err != nil {
        return nil, errors.Wrap(err, "sendStarted")
    }

    resp, err := http.Get(req)
    // Resend the GET request several times until we receive a response
    for i := 0; err != nil; resp, err = http.Get(req) {
        if err, ok := err.(*url.Error); !ok || i > reqRetry {
            return nil, errors.Wrap(err, "sendStarted")
        }
        i++
    }

    // Unmarshal tracker response to get details and list of peers
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return nil, errors.Wrap(err, "sendStarted")
    }

    if resp.StatusCode != 200 {
        return nil, errors.Wrapf(ErrBadStatusCode, "sendStarted: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
    }

    // Update tracker information
    tr.Interval = trResp.Interval
    tr.Complete = trResp.Complete
    tr.Incomplete = trResp.Incomplete

    peersBytes := []byte(trResp.Peers)

    return peer.Unmarshal(peersBytes)
}

func (tr *Tracker) sendStopped(infoHash [20]byte, peerID [20]byte, port uint16, left int) error {
    req, err := tr.buildURL(infoHash, peerID, port, left, "stopped")
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }

    resp, err := http.Get(req)
    // Resend the GET request several times until we receive a response
    for i := 0; err != nil; resp, err = http.Get(req) {
        if err, ok := err.(*url.Error); !ok || i > reqRetry {
            return errors.Wrap(err, "sendStopped")
        }
        i++
    }

    // Unmarshal tracker response to get details
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }

    if resp.StatusCode != 200 {
        return errors.Wrapf(ErrBadStatusCode, "sendStopped: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
    }

    // Update tracker information
    tr.Interval = trResp.Interval
    tr.Complete = trResp.Complete
    tr.Incomplete = trResp.Incomplete

    return nil
}
