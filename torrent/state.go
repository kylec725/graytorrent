package torrent

// State represents the possible states a torrent can be in
type State uint8

// Possible states
const (
    Started  State = 0  // Torrent is downloading and has peers
    Stopped  State = 1  // Torrent is not attempting to download
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

// Command represents commands to switch a torrent's state
type Command uint8

// Possible commands
const (
    Start    Command = 0
    Stop     Command = 1
    Delete   Command = 2
    Shutdown Command = 3
)

func (command Command) String() string {
    switch command {
    case Start:
        return "Start"
    case Stop:
        return "Stop"
    case Delete:
        return "Delete"
    case Shutdown:
        return "Shutdown"
    default:
        return "Unknown"
    }
}
