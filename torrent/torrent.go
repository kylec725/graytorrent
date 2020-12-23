/*
Package torrent provides a library for reading from a torrent file
and storing desired information for leeching or seeding.
Tracker file to handle grabbing information about current
peers and the state of the file.
Write file to handle writing and getting pieces, as well as verifying
the hash of received pieces.
Will communicate with the peers package for sending and receiving
pieces of the torrent.
*/
package torrent

import (
	// "fmt"
)

// Torrent stores metainfo and current progress on a torrent
type Torrent int

func (to Torrent) String() string {
	return "This is a torrent"
}
