/*
Package connect provides functions to form connections or to
check for available network resources
*/
package connect

import (
    "net"
    "os"
    "strconv"
    
    errors "github.com/pkg/errors"
)

// Errors
var (
    ErrBadPortRange = errors.New("Bad port range")
    ErrNoOpenPort = errors.New("Open port not found")
)

// OpenPort finds a local unused port within a range
func OpenPort(portRange []int) (uint16, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return 0, errors.Wrap(err, "GetOpenPort")
    }
    if len(portRange) < 2 {
        return 0, errors.Wrap(ErrBadPortRange, "GetOpenPort")
    }
    for port := portRange[0]; port <= portRange[1]; port++ {
        _, err := net.Dial("tcp", hostname + ":" + strconv.Itoa(port))
        if err != nil {
            return uint16(port), nil
        }
    }
    return 0, errors.Wrap(ErrNoOpenPort, "GetOpenPort")
}
