package torrent

import (
	"math/rand"
	"time"

	"github.com/kylec725/graytorrent/peer/message"
)

func (to *Torrent) unchokeAlg() {
	// Unchoke the peers that are uploading the most
	msgUnchoke := message.Unchoke()
	highRates := to.bestRates()
	prevIndex := -1
	for _, index := range highRates {
		if to.Peers[index].AmChoking && index != prevIndex { // Make sure we don't unchoke the same peer again
			to.Peers[index].AmChoking = false
			to.Peers[index].Send <- (msgUnchoke)
		}
		prevIndex = index
	}

	// Choke peers with lower rates
	msgChoke := message.Choke()
	for i := range to.Peers {
		if !to.Peers[i].AmChoking && !numInSlice(i, highRates) {
			to.Peers[i].AmChoking = true
			to.Peers[i].Send <- (msgChoke)
		}
	}
}

func numInSlice(index int, nums []int) bool {
	for i := range nums {
		if index == nums[i] {
			return true
		}
	}
	return false
}

func (to *Torrent) bestRates() []int {
	highRates := make([]int, 4)
	// Find peers with top 4 download rates
	for i, peer := range to.Peers {
		if peer.DownRate() > to.Peers[highRates[0]].DownRate() && peer.PeerInterested {
			highRates[3] = highRates[2]
			highRates[2] = highRates[1]
			highRates[1] = highRates[0]
			highRates[0] = i
		} else if peer.DownRate() > to.Peers[highRates[1]].DownRate() && peer.PeerInterested {
			highRates[3] = highRates[2]
			highRates[2] = highRates[1]
			highRates[1] = i
		} else if peer.DownRate() > to.Peers[highRates[2]].DownRate() && peer.PeerInterested {
			highRates[3] = highRates[1]
			highRates[2] = i
		} else if peer.DownRate() > to.Peers[highRates[3]].DownRate() && peer.PeerInterested {
			highRates[3] = i
		}
	}
	return highRates
}

func (to *Torrent) optimisticUnchoke() int {
	rand.Seed(time.Now().UnixNano())
	peer := rand.Intn(len(to.Peers))
	return peer
}
