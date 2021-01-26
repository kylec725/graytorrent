package main

import (
	"os"
    "net"

    "github.com/kylec725/graytorrent/torrent"
    flag "github.com/spf13/pflag"
    log "github.com/sirupsen/logrus"
    viper "github.com/spf13/viper"
)

const logLevel = log.DebugLevel  // InfoLevel || DebugLevel || TraceLevel

var (
    err error
    logFile *os.File

    // Flags
    filename string
    verbose bool
    port uint16

    torrentList []torrent.Torrent
    listener net.Listener
)

func init() {
    flag.StringVarP(&filename, "file", "f", "", "Filename of torrent file")
    flag.BoolVarP(&verbose, "verbose", "v", false, "Print events to stdout")
    flag.Parse()

    setupLog()
    log.Info("Graytorrent started")

    setupViper()
    viper.WatchConfig()

    setupListen()
}

func main() {
    defer logFile.Close()
    defer log.Info("Graytorrent stopped")
    defer listener.Close()
    defer saveTorrents()

    go peerListen()  // Listen for incoming peer connections

    // Single file torrent then exit
    if filename != "" {
        singleTorrent()
        return
    }

    // Initialize GUI
    // defer g.Close()

    // Send torrent stopped messages
}
