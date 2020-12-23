package torrent

import (
	// "fmt"
)

// Tracker stores information about a torrent tracker
type Tracker int

func (tr Tracker) String() string {
	return "This is a tracker"
}
