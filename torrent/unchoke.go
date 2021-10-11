package torrent

import (
	"math/rand"
	"time"

	"github.com/kylec725/graytorrent/internal/peer"
	"github.com/kylec725/graytorrent/internal/peer/message"
)

func (to *Torrent) unchokeAlg() {
	msgUnchoke := message.Unchoke()
	msgChoke := message.Choke()
	bestDownloaders := to.bestDownloaders()

	// Loop through all the peers, unchoke them if they are the optimistic unchoke or have a high rate of upload to us
	for i := range to.Peers {
		shouldChoke := true
		for _, peerPtr := range bestDownloaders {
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

// bestDownloaders gets a list of the downloaders with the highest upload rate to us
func (to *Torrent) bestDownloaders() []*peer.Peer {
	bestDownloaders := make([]*peer.Peer, 4)
	// Find peers with top 4 download rates
	for i := range to.Peers {
		if to.Peers[i].PeerInterested && (bestDownloaders[0] == nil || to.Peers[i].DownRate() > bestDownloaders[0].DownRate()) {
			bestDownloaders[3] = bestDownloaders[2]
			bestDownloaders[2] = bestDownloaders[1]
			bestDownloaders[1] = bestDownloaders[0]
			bestDownloaders[0] = &to.Peers[i]
		} else if to.Peers[i].PeerInterested && (bestDownloaders[1] == nil || to.Peers[i].DownRate() > bestDownloaders[1].DownRate()) {
			bestDownloaders[3] = bestDownloaders[2]
			bestDownloaders[2] = bestDownloaders[1]
			bestDownloaders[1] = &to.Peers[i]
		} else if to.Peers[i].PeerInterested && (bestDownloaders[2] == nil || to.Peers[i].DownRate() > bestDownloaders[2].DownRate()) {
			bestDownloaders[3] = bestDownloaders[1]
			bestDownloaders[2] = &to.Peers[i]
		} else if to.Peers[i].PeerInterested && (bestDownloaders[3] == nil || to.Peers[i].DownRate() > bestDownloaders[3].DownRate()) {
			bestDownloaders[3] = &to.Peers[i]
		}
	}
	return bestDownloaders
}

func (to *Torrent) changeOptimisticUnchoke(lastOpUnchoke *time.Time) {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(to.Peers))
	*lastOpUnchoke = time.Now()
	to.optimisticUnchoke = &to.Peers[index]
}
