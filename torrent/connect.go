package torrent

import (
    "net"
    "os"
    "strconv"
    "errors"
)

const basePort = 6881
const portRange = 8

// Finds a local unused port within a range
func getOpenPort() (int, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return 0, err
    }
    for i, port := 0, basePort; i < portRange + 1; i++ {
        _, err := net.Dial("tcp", hostname + ":" + strconv.Itoa(port + i))
        if err != nil {
            return port + i, nil
        }
    }
    return 0, errors.New("Open port not found")
}
