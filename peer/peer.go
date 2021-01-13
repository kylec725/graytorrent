/*
Package peer provides the ability to setup connections with peers and
manage sending and receiving torrent pieces with those peers
*/
package peer

import (
    "net"
)

// Peer stores info about connecting to peers as well as their state
type Peer struct {
    addr net.IP
    port uint16
}
