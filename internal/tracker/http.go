package tracker

import (
	"fmt"
	"net/url"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
	"github.com/kylec725/gray/internal/common"
	"github.com/kylec725/gray/internal/peer"
	"github.com/pkg/errors"
)

// Errors
var (
	ErrBadStatusCode = errors.New("Expected status code 200")
)

type bencodeTrackerResp struct {
	Interval   int    `bencode:"interval"`
	Peers      string `bencode:"peers"`
	Failure    string `bencode:"failure reason"`
	Complete   int    `bencode:"complete"`
	Incomplete int    `bencode:"incomplete"`
}

func (tr Tracker) buildURL(event string, info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) (string, error) {
	base, err := url.Parse(tr.Announce)
	if err != nil {
		return "", errors.Wrap(err, "buildURL")
	}

	params := url.Values{
		"info_hash":  []string{string(info.InfoHash[:])},
		"peer_id":    []string{string(info.PeerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{strconv.Itoa(uploaded)},
		"downloaded": []string{strconv.Itoa(downloaded)},
		"left":       []string{strconv.Itoa(left)},
		"compact":    []string{"1"},
		"event":      []string{event},
		"numwant":    []string{strconv.Itoa(numWant)},
	}

	if event == "" {
		delete(params, "event")
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

// NOTE: still get a connection reset by peer error from some trackers
func (tr *Tracker) httpStarted(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]peer.Peer, error) {
	// Request
	req, err := tr.buildURL("started", info, port, uploaded, downloaded, left)
	if err != nil {
		return nil, errors.Wrap(err, "httpStarted")
	}

	resp, err := tr.httpClient.Get(req)
	if err != nil {
		unwrapped := errors.Unwrap(err)
		// Retry again if connection was reset
		if unwrapped != nil && unwrapped.Error() == "read: connection reset by peer" {
			fmt.Println("connection was reset but we good")
			resp, err = tr.httpClient.Get(req)
			if err != nil {
				return nil, errors.Wrap(err, "httpStarted")
			}
		} else {
			return nil, errors.Wrap(err, "httpStarted")
		}
	}
	defer resp.Body.Close()

	// Unmarshal tracker response to get details and list of peers
	var trResp bencodeTrackerResp
	err = bencode.Unmarshal(resp.Body, &trResp)
	if err != nil {
		return nil, errors.Wrap(err, "httpStarted")
	}

	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(ErrBadStatusCode, "httpStarted: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
	}

	// Update tracker information
	tr.Interval = trResp.Interval

	// Get peer information
	peersBytes := []byte(trResp.Peers)
	peersList, err := peer.Unmarshal(peersBytes, info)
	return peersList, errors.Wrap(err, "httpStarted")
}

func (tr *Tracker) httpStopped(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) error {
	// Request
	req, err := tr.buildURL("stopped", info, port, uploaded, downloaded, left)
	if err != nil {
		return errors.Wrap(err, "httpStopped")
	}

	resp, err := tr.httpClient.Get(req)
	if err != nil {
		return errors.Wrap(err, "httpStopped")
	}
	defer resp.Body.Close()

	// Unmarshal tracker response to get details
	var trResp bencodeTrackerResp
	err = bencode.Unmarshal(resp.Body, &trResp)
	if err != nil {
		return errors.Wrap(err, "httpStopped")
	}

	if resp.StatusCode != 200 {
		return errors.Wrapf(ErrBadStatusCode, "httpStopped: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
	}

	// Update tracker information
	tr.Interval = trResp.Interval

	return nil
}

func (tr *Tracker) httpCompleted(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) error {
	// Request
	req, err := tr.buildURL("completed", info, port, uploaded, downloaded, left)
	if err != nil {
		return errors.Wrap(err, "httpCompleted")
	}

	resp, err := tr.httpClient.Get(req)
	if err != nil {
		return errors.Wrap(err, "httpCompleted")
	}
	defer resp.Body.Close()

	// Unmarshal tracker response to get details
	var trResp bencodeTrackerResp
	err = bencode.Unmarshal(resp.Body, &trResp)
	if err != nil {
		return errors.Wrap(err, "httpCompleted")
	}

	if resp.StatusCode != 200 {
		return errors.Wrapf(ErrBadStatusCode, "httpCompleted: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
	}

	// Update tracker information
	tr.Interval = trResp.Interval

	return nil
}

func (tr *Tracker) httpAnnounce(info *common.TorrentInfo, port uint16, uploaded, downloaded, left int) ([]peer.Peer, error) {
	// Request
	req, err := tr.buildURL("", info, port, uploaded, downloaded, left)
	if err != nil {
		return nil, errors.Wrap(err, "httpAnnounce")
	}

	resp, err := tr.httpClient.Get(req)
	if err != nil {
		return nil, errors.Wrap(err, "httpAnnounce")
	}
	defer resp.Body.Close()

	// Unmarshal tracker response to get details
	var trResp bencodeTrackerResp
	err = bencode.Unmarshal(resp.Body, &trResp)
	if err != nil {
		return nil, errors.Wrap(err, "httpAnnounce")
	}

	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(ErrBadStatusCode, "httpAnnounce: GET status code %d and reason '%s'", resp.StatusCode, trResp.Failure)
	}

	// Update tracker information
	tr.Interval = trResp.Interval

	// Get peer information
	peersBytes := []byte(trResp.Peers)
	peersList, err := peer.Unmarshal(peersBytes, info)
	return peersList, errors.Wrap(err, "httpStarted")
}
