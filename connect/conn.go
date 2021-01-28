/*
Package connect provides functions to form connections or to
check for available network resources.
*/
package connect

import (
    "net"
    "time"
    "io"
    "encoding/binary"
    "context"
    "fmt"
    
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
)

const pollTimeout = 3 * time.Second

// Errors
var (
    ErrTimeout = errors.New("Connection operation timed out")
    ErrSend = errors.New("Unexpected number of bytes sent")
)

// Conn is a wrapper around net.Conn with a variable timeout for read/write calls
type Conn struct {
    Conn net.Conn
    Timeout time.Duration
}

// Write sends data over a connection, returns an error if not all of the data is sent
func (conn *Conn) Write(buf []byte) (int, error) {
    defer conn.Conn.SetWriteDeadline(time.Time{})  // Reset timeout
    conn.Conn.SetWriteDeadline(time.Now().Add(conn.Timeout))
    bytesSent, err := conn.Conn.Write(buf)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return bytesSent, errors.Wrap(ErrTimeout, "Write")
        }
        return bytesSent, errors.Wrap(err, "Write")
    } else if bytesSent != len(buf) {
        return bytesSent, errors.Wrap(ErrSend, "Write")
    }
    return bytesSent, nil
}

// Read reads in data from a connection, returns an error if the buffer is not filled
func (conn *Conn) Read(buf []byte) (int, error) {
    defer conn.Conn.SetReadDeadline(time.Time{})  // Reset timeout
    conn.Conn.SetReadDeadline(time.Now().Add(conn.Timeout))
    bytesRead, err := io.ReadFull(conn.Conn, buf)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return bytesRead, errors.Wrap(ErrTimeout, "Read")
        }
    }
    return bytesRead, errors.Wrap(err, "Read")
}

// Close closes a connection
func (conn *Conn) Close() error {
    return conn.Conn.Close()
}

func (conn *Conn) pollRead(buf []byte) (int, error) {
    defer conn.Conn.SetReadDeadline(time.Time{})  // Reset timeout
    conn.Conn.SetReadDeadline(time.Now().Add(pollTimeout))
    bytesRead, err := io.ReadFull(conn.Conn, buf)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return bytesRead, errors.Wrap(ErrTimeout, "Read")
        }
    }
    return bytesRead, errors.Wrap(err, "Read")
}

// Poll scans a connection for data and returns it over a channel
func (conn *Conn) Poll(ctx context.Context, output chan []byte) {
    // Cleanup
    defer func() {
        conn.Conn.Close()
        close(output)
    }()
    for {
        select {
        case <-ctx.Done():
            return
        default:
            buf := make([]byte, 4)  // Expect message length prefix of 4 bytes
            if _, err := conn.pollRead(buf); err != nil {
                if errors.Is(err, ErrTimeout) {  // Don't terminate if we don't receive anything
                    continue
                }
                log.WithFields(log.Fields{"peer": conn.Conn.RemoteAddr().String(), "error": err.Error()}).Debug("Receiving message length failed")
                return
            }
            length := binary.BigEndian.Uint32(buf)
            if length == 0 {  // Keep-alive
                output <- make([]byte, 0)
                continue
            }
            fmt.Println("message length:", length)

            // Message
            buf = make([]byte, length)
            if _, err := conn.Read(buf); err != nil {
                log.WithFields(log.Fields{"peer": conn.Conn.RemoteAddr().String(), "error": err.Error()}).Debug("Receiving message payload failed")
                return
            }
            output <- buf
        }
    }
}
