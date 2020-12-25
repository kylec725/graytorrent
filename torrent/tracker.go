package torrent

import (
    // "fmt"
)

// Tracker stores information about a torrent tracker
type Tracker struct {
    Announce string
    Status string
    Interval int
    ID string
}

func (tr Tracker) String() string {
    return "This is a tracker"
}
