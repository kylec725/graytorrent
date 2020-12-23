package torrent

import (
	"fmt"
)

// Torrent stores metainfo and current progress on a torrent
type Torrent int

// Print the torrent in a human-readable form
func (to Torrent) Print() {
	fmt.Println("This is the torrents file")
}
