package torrent

// State represents the possible states a torrent can be in
type State uint8

// Possible states for torrents
const (
	Started  State = iota // Torrent is downloading and has peers
	Stopped               // Torrent is not complete nor attempting to download
	Stalled               // Torrent is attempting to download, but has no peers
	Seeding               // Torrent is complete and seeding
	Complete              // Torrent is complete and not seeding
)

func (state State) String() string {
	switch state {
	case Started:
		return "Started"
	case Stopped:
		return "Stopped"
	case Stalled:
		return "Stalled"
	case Seeding:
		return "Seeding"
	case Complete:
		return "Complete"
	default:
		return "Unknown"
	}
}
