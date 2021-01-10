package torrent

import (
    "net"
    "os"
    "strconv"
    "errors"

    viper "github.com/spf13/viper"
)

// Finds a local unused port within a range
func getOpenPort() (int, error) {
    viper := viper.GetViper()
    portRange := viper.GetIntSlice("network.portrange")

    hostname, err := os.Hostname()
    if err != nil {
        return 0, err
    }
    for port := portRange[0]; port <= portRange[1]; port++ {
        _, err := net.Dial("tcp", hostname + ":" + strconv.Itoa(port))
        if err != nil {
            return port, nil
        }
    }
    return 0, errors.New("Open port not found")
}
