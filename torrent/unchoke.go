package torrent

import "github.com/kylec725/graytorrent/peer/message"

func (to *Torrent) unchokeAlg() {
	msgUnchoke := message.Unchoke()
	// Choke all peers
	for i := range to.Peers {
		if !to.Peers[i].AmChoking {
			to.Peers[i].SendMessage(msgUnchoke)
		}
	}
	highRates := to.bestRates()
	msgChoke := message.Choke()
	// Unchoke the peers with the top 4 rates
	prevIndex := highRates[0]
	for _, index := range highRates {
		if index != prevIndex { // Make sure we don't unchoke the same peer again
			to.Peers[index].SendMessage(msgChoke)
		}
		prevIndex = index
	}
}

func (to *Torrent) bestRates() [4]int {
	var highRates [4]int
	// Find peers with top 4 download rates
	for i, peer := range to.Peers {
		if peer.Rate > to.Peers[highRates[0]].Rate {
			highRates[3] = highRates[2]
			highRates[2] = highRates[1]
			highRates[1] = highRates[0]
			highRates[0] = i
		} else if peer.Rate > to.Peers[highRates[1]].Rate {
			highRates[3] = highRates[2]
			highRates[2] = highRates[1]
			highRates[1] = i
		} else if peer.Rate > to.Peers[highRates[2]].Rate {
			highRates[3] = highRates[1]
			highRates[2] = i
		} else if peer.Rate > to.Peers[highRates[3]].Rate {
			highRates[3] = i
		}
	}
	return highRates
}
