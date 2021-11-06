package common

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type contextKey string // Custom key types are recommended

// Context keys
var (
	KeyPort = contextKey("port")
)

// Port returns the port number from the current context
func Port(ctx context.Context) uint16 {
	port, ok := ctx.Value(KeyPort).(uint16)
	if !ok {
		log.Fatal("Failed to get port from context")
	}
	return port
}
