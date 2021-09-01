package common

// State indicates the state and thus behavior of a torrent goroutine
type State uint8

// Possible states for torrents
const (
	Started State = iota
	Stopped
)
