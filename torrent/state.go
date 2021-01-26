package torrent

// State represents the possible states a torrent can be in
type State uint8

// Possible states
const (
    Started  State = 0  // Torrent is downloading and has peers
    Stopped  State = 1  // Torrent is not complete nor attempting to download
    Stalled  State = 2  // Torrent is attempting to download, but has no peers
    Seeding  State = 3  // Torrent is complete and seeding
    Complete State = 4  // Torrent is complete and not seeding
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
