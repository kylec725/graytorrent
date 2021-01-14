package torrent

import (
    "strconv"
    "errors"
    "net/http"
    "net/url"
    "fmt"

    "github.com/kylec725/graytorrent/peer"
    bencode "github.com/jackpal/bencode-go"
)

const reqRetry = 5

func (tr *Tracker) getPeers(infoHash [20]byte, peerID [20]byte, port uint16, left int) ([]peer.Peer, error) {
    req, err := tr.buildURL(infoHash, peerID, port, left, "started")
    if err != nil {
        return nil, err
    }

    resp, err := http.Get(req)
    // Resend the GET request several times until we receive a response
    for i := 0; err != nil; resp, err = http.Get(req) {
        if err, ok := err.(*url.Error); !ok {
            return nil, err
        } else if i++; i > reqRetry {
            return nil, errors.New("Sent GET requests " + strconv.Itoa(reqRetry) + " times with no response")
        }
    }

    // Unmarshal tracker response to get new interval and list of peers
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != 200 {
        return nil, errors.New("Got status code " + strconv.Itoa(resp.StatusCode) + ": " + trResp.Failure)
    }


    tr.Interval = trResp.Interval
    peersBytes := []byte(trResp.Peers)

    fmt.Println("completed:", trResp.Complete)
    fmt.Println("incompleted:", trResp.Incomplete)

    return peer.Unmarshal(peersBytes)
}

func (tr *Tracker) sendStopped(infoHash [20]byte, peerID [20]byte, port uint16, left int) error {
    req, err := tr.buildURL(infoHash, peerID, port, left, "stopped")
    if err != nil {
        return err
    }

    resp, err := http.Get(req)
    // Resend the GET request several times until we receive a response
    for i := 0; err != nil; resp, err = http.Get(req) {
        if err, ok := err.(*url.Error); !ok {
            return err
        } else if i++; i > reqRetry {
            return errors.New("Sent GET requests " + strconv.Itoa(reqRetry) + " times with no response")
        }
    }

    // Unmarshal tracker response
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    resp.Body.Close()
    if err != nil {
        return err
    }

    if resp.StatusCode != 200 {
        return errors.New("Got status code " + strconv.Itoa(resp.StatusCode) + ": " + trResp.Failure)
    }

    tr.Interval = trResp.Interval

    return nil
}
