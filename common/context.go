package common

import (
    "context"
)

type contextKey string  // Custom key types are recommended

// Context keys
var (
    KeyInfo = contextKey("info")
    KeyPort = contextKey("port")
)

// Info returns a TorrentInfo object from the current context
func Info(ctx context.Context) TorrentInfo {
    info, ok := ctx.Value(KeyInfo).(*TorrentInfo)  // Use a pointer so that we get updated info values
    if !ok {
        panic("Failed to get TorrentInfo from context")
    }
    return *info
}

// Port returns the port number from the current context
func Port(ctx context.Context) uint16 {
    port, ok := ctx.Value(KeyPort).(uint16)
    if !ok {
        panic("Failed to get port from context")
    }
    return port
}
