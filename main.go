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
    torrentList []torrent.Torrent
    listener net.Listener
    filename string
    verbose bool
    port uint16
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
        to, err := newTorrent(filename, port)
        if err != nil {
            fmt.Println("Single torrent failed:", err)
            return
        }
        to.Start()
        fmt.Println("Torrent done:", to.Info.Name)
        return
    }

    // Initialize GUI
    // defer g.Close()

    // Send torrent stopped messages
    // Save torrent progresses to history file
}

func newTorrent(filename string, port uint16) (torrent.Torrent, error) {
    to := torrent.Torrent{Path: filename, Port: port}
    if err := to.Setup(); err != nil {
        log.WithFields(log.Fields{"file": filename, "error": err.Error()}).Info("Torrent setup failed")
        return torrent.Torrent{}, err
    }
    torrentList = append(torrentList, to)
    log.WithField("name", to.Info.Name).Info("Torrent added")
    return to, nil
}
