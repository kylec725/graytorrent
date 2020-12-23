package torrent

import (
	"fmt"
)

// Tracker stores information about a torrent tracker
type Tracker int

// Print the torrent in a human-readable form
func (tr Tracker) Print() {
	fmt.Println("This is the tracker file")
}
