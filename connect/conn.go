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
    "fmt"
    
    "github.com/pkg/errors"
)

const retry = 3

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
    Timeout time.Duration
}

// Write sends data over a connection, returns an error if not all of the data is sent
func (conn *Conn) Write(buf []byte) error {
    conn.Conn.SetWriteDeadline(time.Time{})  // No deadline for writing
    var err error
    for i := 0; i < retry; i++ {
        _, err = conn.Conn.Write(buf)
        if err == nil {
            break
        }
        fmt.Println("write error unwrapped:", errors.Unwrap(err))
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return errors.Wrap(ErrTimeout, "Write")
        } else if errors.Unwrap(err).Error() == "connection reset by peer" {
            fmt.Println("Write caught: connection reset by peer")
            continue
        } else if errors.Unwrap(err).Error() == "use of closed network connection" {
            fmt.Println("Write caught: use of closed network connection")
            continue
        }
        return errors.Wrap(err, "Write")
    }
    return errors.Wrap(err, "Write")
}

// Read reads in data from a connection, returns an error if the buffer is not filled
func (conn *Conn) Read(buf []byte) error {
    err := conn.Conn.SetReadDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "Read")
    }
    for i := 0; i < retry; i++ {
        _, err = conn.Conn.Read(buf)
        if err == nil {
            break
        }
        fmt.Println("read error unwrapped:", errors.Unwrap(err))
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return errors.Wrap(ErrTimeout, "Read")
        }
        // } else if errors.Unwrap(err).Error() == "read: connection reset by peer" {
        //     fmt.Println("Read caught: connection reset by peer")
        //     continue
        // }
        return errors.Wrap(err, "Read")
    }
    return errors.Wrap(err, "Read")
}

// ReadFull reads until the buffer is full
func (conn *Conn) ReadFull(buf []byte) error {
    err := conn.Conn.SetReadDeadline(time.Now().Add(conn.Timeout))
    if err != nil {
        return errors.Wrap(err, "ReadFull")
    }
    _, err = io.ReadFull(conn.Conn, buf)
    return errors.Wrap(err, "ReadFull")
}

// Close closes a connection
func (conn *Conn) Close() error {
    return conn.Conn.Close()
}

// Await polls a connection for data and returns it over a channel
func (conn *Conn) Await(output chan []byte) {
    for {
        if err := conn.Conn.SetReadDeadline(time.Now().Add(conn.Timeout)); err != nil {  // We use this to check if the connection was closed
            goto exit
        }

        buf := make([]byte, 4)  // Expect message length prefix of 4 bytes
        if bytesRead, err := io.ReadFull(conn.Conn, buf); err != nil {
            goto exit
        } else if bytesRead != 4 {
            goto exit
        }
        length := binary.BigEndian.Uint32(buf)
        buf = make([]byte, length)
        if bytesRead, err := io.ReadFull(conn.Conn, buf); err != nil {
            goto exit
        } else if uint32(bytesRead) != length {
            goto exit
        }
        output <- buf
    }

    exit:
    close(output)
    conn.Conn.Close()
    return
}
