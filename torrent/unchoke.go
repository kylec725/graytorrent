package torrent

import (
	"math/rand"
	"time"

	"github.com/kylec725/graytorrent/peer"
	"github.com/kylec725/graytorrent/peer/message"
)

func (to *Torrent) unchokeAlg() {
	msgUnchoke := message.Unchoke()
	msgChoke := message.Choke()
	highRates := to.highRatePeers()

	// Loop through all the peers, unchoke them if they are the optimistic unchoke or have a high rate of upload to us
	for i := range to.Peers {
		shouldChoke := true
		for _, peerPtr := range highRates {
			if &to.Peers[i] == peerPtr && peerPtr.AmChoking { // Make sure we don't unchoke the same peer again
				shouldChoke = false
				to.Peers[i].AmChoking = false
				to.Peers[i].Send <- (msgUnchoke)
			}
		}
		if &to.Peers[i] == to.optimisticUnchoke {
			shouldChoke = false
			if to.Peers[i].AmChoking {
				to.Peers[i].AmChoking = false
				to.Peers[i].Send <- (msgUnchoke)
			}
		}
		// Choke the peer if we aren't already choking them
		if shouldChoke && !to.Peers[i].AmChoking {
			to.Peers[i].AmChoking = true
			to.Peers[i].Send <- (msgChoke)
		}
	}
}

func (to *Torrent) highRatePeers() []*peer.Peer {
	highRates := make([]*peer.Peer, 4)
	// Find peers with top 4 download rates
	for i := range to.Peers {
		if to.Peers[i].PeerInterested && (highRates[0] == nil || to.Peers[i].DownRate() > highRates[0].DownRate()) {
			highRates[3] = highRates[2]
			highRates[2] = highRates[1]
			highRates[1] = highRates[0]
			highRates[0] = &to.Peers[i]
		} else if to.Peers[i].PeerInterested && (highRates[1] == nil || to.Peers[i].DownRate() > highRates[1].DownRate()) {
			highRates[3] = highRates[2]
			highRates[2] = highRates[1]
			highRates[1] = &to.Peers[i]
		} else if to.Peers[i].PeerInterested && (highRates[2] == nil || to.Peers[i].DownRate() > highRates[2].DownRate()) {
			highRates[3] = highRates[1]
			highRates[2] = &to.Peers[i]
		} else if to.Peers[i].PeerInterested && (highRates[3] == nil || to.Peers[i].DownRate() > highRates[3].DownRate()) {
			highRates[3] = &to.Peers[i]
		}
	}
	return highRates
}

func (to *Torrent) changeOptimisticUnchoke(lastOpUnchoke *time.Time) {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(to.Peers))
	*lastOpUnchoke = time.Now()
	to.optimisticUnchoke = &to.Peers[index]
}
