/*
Package connect provides functions to form connections or to
check for available network resources.
*/
package connect

import (
    "net"
    "time"
    "io"
    
    "github.com/pkg/errors"
)

const readFullRetry = 10

// Errors
var (
    ErrTimeout = errors.New("Connection operation timed out")
    ErrSend = errors.New("Unexpected number of bytes sent")
    ErrRcv = errors.New("Unexpected number of bytes received")
    ErrReadFull = errors.New("Retried connection reading too many times")
)

// Conn is a wrapper around net.Conn with a variable timeout for read/write calls
type Conn struct {
    Conn net.Conn
    Timeout time.Duration  // TODO remove timeout in favor of goroutine based receiving
    shutdown bool
}

// Write sends data over a connection, returns an error if not all of the data is sent
func (conn *Conn) Write(buf []byte) error {
    err := conn.Conn.SetDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "Write")
    }
    bytesSent, err := conn.Conn.Write(buf)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return errors.Wrap(ErrTimeout, "Write")
        }
        return errors.Wrap(err, "Write")
    } else if bytesSent != len(buf) {
        return errors.Wrap(ErrSend, "Write")
    }
    return nil
}

// Read reads in data from a connection, returns an error if the buffer is not filled
func (conn *Conn) Read(buf []byte) error {
    err := conn.Conn.SetDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "Read")
    }
    bytesRead, err := conn.Conn.Read(buf)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return errors.Wrap(ErrTimeout, "Read")
        }
        return errors.Wrap(err, "Read")
    } else if bytesRead != len(buf) {
        return errors.Wrap(ErrRcv, "Read")
    }
    return nil
}

// ReadFull reads until the buffer is full
func (conn *Conn) ReadFull(buf []byte) error {
    err := conn.Conn.SetDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "ReadFull")
    }
    _, err = io.ReadFull(conn.Conn, buf)
    return errors.Wrap(err, "ReadFull")
}

// Close closes a connection
func (conn *Conn) Close() error {
    conn.shutdown = true
    return conn.Conn.Close()
}

// Await polls a connection for data and returns it over a channel
// TODO
func (conn *Conn) Await(output chan []byte) {
    conn.shutdown = false
    if err := conn.Conn.SetDeadline(time.Time{}); err != nil {
        close(output)
        return
    }
    for {
        // Exit point
        if conn.shutdown {
            close(output)
            return
        }
        buf := make([]byte, 1)
        if bytesRead, err := conn.Conn.Read(buf); err != nil {
            conn.shutdown = true
            continue
        } else if bytesRead != 1 {
            conn.shutdown = true
            continue
        }
        length := int(buf[0])
        buf = make([]byte, length)
        if bytesRead, err := io.ReadFull(conn.Conn, buf); err != nil {
            conn.shutdown = true
            continue
        } else if bytesRead != length {
            conn.shutdown = true
            continue
        }
        output <- buf
    }
}
