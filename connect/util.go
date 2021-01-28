package connect

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Errors
var (
	ErrBadPortRange = errors.New("Bad port range")
	ErrNoOpenPort   = errors.New("Open port not found")
)

// OpenPort finds a local unused port within a range
func OpenPort(portRange []int) (uint16, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return 0, errors.Wrap(err, "OpenPort")
	}
	if len(portRange) < 2 {
		return 0, errors.Wrap(ErrBadPortRange, "OpenPort")
	}
	for port := portRange[0]; port <= portRange[1]; port++ {
		_, err := net.Dial("tcp", hostname+":"+strconv.Itoa(port))
		if err != nil {
			return uint16(port), nil
		}
	}
	return 0, errors.Wrap(ErrNoOpenPort, "OpenPort")
}

// PortFromAddr returns the port from a string address
func PortFromAddr(addr string) (uint16, error) {
	split := strings.Split(addr, ":")
	portString := split[len(split)-1]
	port, err := strconv.Atoi(portString)
	if err != nil {
		return 0, errors.Wrap(err, "PortFromAddr")
	}
	return uint16(port), nil
}
