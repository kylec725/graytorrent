package torrent

import (
    "net/http"
    "net/url"

    "github.com/kylec725/graytorrent/peer"
    errors "github.com/pkg/errors"
    bencode "github.com/jackpal/bencode-go"
)

const reqRetry = 5

// Errors
var (
    ErrReqRetry = errors.Errorf("Sent GET requests %d times with no response", reqRetry)
    ErrBadStatusCode = errors.New("Did not get status code 200")
)

func (tr *Tracker) getPeers(infoHash [20]byte, peerID [20]byte, port uint16, left int) ([]peer.Peer, error) {
    req, err := tr.buildURL(infoHash, peerID, port, left, "started")
    if err != nil {
        return nil, errors.Wrap(err, "getPeers")
    }

    resp, err := http.Get(req)
    // Resend the GET request several times until we receive a response
    for i := 0; err != nil; resp, err = http.Get(req) {
        if !errors.Is(err, err.(*url.Error)) {
            return nil, errors.Wrap(err, "getPeers")
        } else if i++; i > reqRetry {
            return nil, errors.Wrap(ErrReqRetry, "getPeers")
        }
    }

    // Unmarshal tracker response to get new interval and list of peers
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return nil, errors.Wrap(err, "getPeers")
    }

    if resp.StatusCode != 200 {
        return nil, errors.Wrap(ErrBadStatusCode, "getPeers")
    }


    tr.Interval = trResp.Interval
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
        if !errors.Is(err, err.(*url.Error)) {
            return errors.Wrap(err, "sendStopped")
        } else if i++; i > reqRetry {
            return errors.Wrap(ErrReqRetry, "sendStopped")
        }
    }

    // Unmarshal tracker response
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }

    if resp.StatusCode != 200 {
        return errors.Wrap(ErrBadStatusCode, "sendStopped")
    }

    tr.Interval = trResp.Interval

    return nil
}
