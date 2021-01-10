/*
Package connect provides functions to form connections or to
check for available network resources
*/
package connect

import (
    "net"
    "os"
    "strconv"
    "errors"
)

// GetOpenPort finds a local unused port within a range
func GetOpenPort(portRange []int) (uint16, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return 0, err
    }
    if len(portRange) < 2 {
        return 0, errors.New("Need port range of two integers")
    }
    for port := portRange[0]; port <= portRange[1]; port++ {
        _, err := net.Dial("tcp", hostname + ":" + strconv.Itoa(port))
        if err != nil {
            return uint16(port), nil
        }
    }
    return 0, errors.New("Open port not found")
}
