package peer

import (
    "net"
)

const protocol = "BitTorrent protocol"

func (peer *Peer) newHandshake() []byte {
    pstr := protocol
    pstrLen := uint8(len(pstr))
    handshake := make([]byte, 49 + pstrLen)
    handshake[0] = pstrLen
    curr := 1
    curr += copy(handshake[curr:], pstr)
    curr += copy(handshake[curr:], peer.torrent.InfoHash)
    curr += copy(handshake[curr:], peer.torrent.PeerID)

}

func (peer *Peer) sendHandshake() error {
    handshake := newHandshake()
    // Start the TCP connection
    conn, err := net.Dial("tcp", peer.String())
    if err != nil {
        return errors.Wrap(err, "sendHandshake")
    }
    conn.SetDeadline(connTimeout)

    // Send the handshake
    byteSent, err := conn.Write(handshake)
    if err != nil {
        return errors.Wrap(err, "sendHandshake")
    } else if bytesSent != len(handshake) {  // TODO probably will change, not sure if all bytes are guaranteed to be sent
        fmt.Println("Fix sendHandshake")
        return errors.New("Unexpected number of bytes sent")
    }

    peer.conn = conn
    return nil
}
