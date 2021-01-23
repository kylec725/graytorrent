/*
Package connect provides functions to form connections or to
check for available network resources
*/
package connect

import (
    "net"
    "time"
    
    "github.com/pkg/errors"
)

// Errors
var (
    ErrTimeout = errors.New("Connection operation timed out")
    ErrSend = errors.New("Unexpected number of bytes sent")
    ErrRcv = errors.New("Unexpected number of bytes received")
)

// Conn is a wrapper around net.Conn with variable timeout for read/write calls
type Conn struct {
    Conn net.Conn
    Timeout time.Duration
}

// Write sends data over a connection, returns an error if not all of the data is sent
func (conn *Conn) Write(data []byte) error {
    err := conn.Conn.SetDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "Write")
    }
    bytesSent, err := conn.Conn.Write(data)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return errors.Wrap(ErrTimeout, "Write")
        }
        return errors.Wrap(err, "Write")
    } else if bytesSent != len(data) {
        return errors.Wrap(ErrSend, "Write")
    }
    return nil
}

// Read reads in data from a connection, returns an error if the buffer is not filled
func (conn *Conn) Read(data []byte) error {
    err := conn.Conn.SetDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "Read")
    }
    bytesRead, err := conn.Conn.Write(data)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return errors.Wrap(ErrTimeout, "Read")
        }
        return errors.Wrap(err, "Read")
    } else if bytesRead != len(data) {
        return errors.Wrap(ErrRcv, "Read")
    }
    return nil
}

// Close closes a connection
func (conn *Conn) Close() error {
    return conn.Conn.Close()
}
