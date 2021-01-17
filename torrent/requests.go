package torrent

import (
    "github.com/kylec725/graytorrent/peer"
    "github.com/pkg/errors"
    bencode "github.com/jackpal/bencode-go"
)

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

    resp, err := tr.httpClient.Get(req)
    if err != nil {
        // Resend a GET request once
        resp, err = tr.httpClient.Get(req)
        if err != nil {
            return nil, errors.Wrap(err, "sendStarted")
        }
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
    peersList, err := peer.Unmarshal(peersBytes)
    if err != nil {
        return nil, errors.Wrap(err, "sendStarted")
    }

    return peersList, nil
}

func (tr *Tracker) sendStopped(infoHash [20]byte, peerID [20]byte, port uint16, left int) error {
    req, err := tr.buildURL(infoHash, peerID, port, left, "stopped")
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }

    resp, err := tr.httpClient.Get(req)
    if err != nil {
        // Resend a GET request once
        resp, err = tr.httpClient.Get(req)
        if err != nil {
            return errors.Wrap(err, "sendStopped")
        }
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
