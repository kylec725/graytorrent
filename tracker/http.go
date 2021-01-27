package tracker

import (
    "net/url"
    "strconv"
    "fmt"

    "github.com/kylec725/graytorrent/common"
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

func (tr Tracker) buildURL(event string, info common.TorrentInfo, port uint16) (string, error) {
    base, err := url.Parse(tr.Announce)
    if err != nil {
        return "", errors.Wrap(err, "buildURL")
    }

    params := url.Values{
        "info_hash": []string{string(info.InfoHash[:])},
        "peer_id": []string{string(info.PeerID[:])},
        "port": []string{strconv.Itoa(int(port))},
        "uploaded": []string{"0"},
        "downloaded": []string{"0"},
        "left": []string{strconv.Itoa(info.Left)},
        "compact": []string{"1"},
        "event": []string{event},
    }

    // if event == "" {
    //     delete(params, "event")
    // }

    base.RawQuery = params.Encode()
    return base.String(), nil
}

func (tr *Tracker) sendStarted(info common.TorrentInfo, port uint16) ([]peer.Peer, error) {
    req, err := tr.buildURL("started", info, port)
    if err != nil {
        return nil, errors.Wrap(err, "sendStarted")
    }

    resp, err := tr.httpClient.Get(req)
    if err != nil {
        unwrapped := errors.Unwrap(err)
        // Retry again if connection was reset
        if unwrapped != nil && unwrapped.Error() == "read: connection reset by peer" {
            fmt.Println("connection was reset but we good")
            resp, err = tr.httpClient.Get(req)
            if err != nil {
                return nil, errors.Wrap(err, "sendStarted")
            }
        } else {
            return nil, errors.Wrap(err, "sendStarted")
        }
    }
    defer resp.Body.Close()

    // Unmarshal tracker response to get details and list of peers
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    if err != nil {
        return nil, errors.Wrap(err, "sendStarted")
    }

    if resp.StatusCode != 200 {
        return nil, errors.Wrapf(ErrBadStatusCode, "sendStarted: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
    }

    // Update tracker information
    tr.Interval = trResp.Interval

    peersBytes := []byte(trResp.Peers)
    peersList, err := peer.Unmarshal(peersBytes, info)
    if err != nil {
        return nil, errors.Wrap(err, "sendStarted")
    }

    return peersList, nil
}

func (tr *Tracker) sendStopped(info common.TorrentInfo, port uint16) error {
    req, err := tr.buildURL("stopped", info, port)
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }

    resp, err := tr.httpClient.Get(req)
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }
    defer resp.Body.Close()

    // Unmarshal tracker response to get details
    var trResp bencodeTrackerResp
    err = bencode.Unmarshal(resp.Body, &trResp)
    if err != nil {
        return errors.Wrap(err, "sendStopped")
    }

    if resp.StatusCode != 200 {
        return errors.Wrapf(ErrBadStatusCode, "sendStopped: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
    }

    // Update tracker information
    tr.Interval = trResp.Interval

    return nil
}
