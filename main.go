package main

import (
	"os"
    "fmt"
    "net"

    "github.com/kylec725/graytorrent/torrent"
    flag "github.com/spf13/pflag"
    log "github.com/sirupsen/logrus"
    viper "github.com/spf13/viper"
)

const logLevel = log.TraceLevel  // InfoLevel || DebugLevel || TraceLevel

var (
    err error
    logFile *os.File

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

    setupLog(verbose)
    log.Info("Graytorrent started")

    setupViper()
    viper.WatchConfig()

    listener, port = setupListen()
}

func main() {
    defer logFile.Close()
    defer log.Info("Graytorrent stopped")

    go peerListen(listener)  // Listen for incoming peer connections

    // Single file torrent then exit
    if filename != "" {
        to, err := addTorrent(torrentList, filename, port)
        if err != nil {
            fmt.Println("Single torrent failed:", err)
            log.WithFields(log.Fields{"filename": filename, "error": err.Error()}).Info("Failed to add torrent")
            return
        }
        log.WithField("name", to.Info.Name).Info("Torrent added")
        to.Start()
        shutdown(torrentList)
        fmt.Println("Torrent done:", to.Info.Name)
        return
    }

    // Initialize GUI
    // defer g.Close()

    // Send torrent stopped messages
    // Save torrent progresses to history file
}
