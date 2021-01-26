package main

import (
	"os"
    "fmt"

    "github.com/kylec725/graytorrent/connect"
    "github.com/kylec725/graytorrent/torrent"
    flag "github.com/spf13/pflag"
    log "github.com/sirupsen/logrus"
    viper "github.com/spf13/viper"
)

const logLevel = log.TraceLevel  // InfoLevel || DebugLevel || TraceLevel

var (
    logFile *os.File
    port uint16
    err error
    filename string
    torrentList []torrent.Torrent
)

func init() {
    flag.StringVarP(&filename, "file", "f", "", "Filename of torrent file")
    flag.Parse()

    setupLog()
    log.Info("Graytorrent started")

    setupViper()
    viper.WatchConfig()
}

func init() {
    portRange := viper.GetIntSlice("network.portrange")
    port, err = connect.OpenPort(portRange)
    if err != nil {
        log.WithFields(log.Fields{
            "portrange": portRange,
            "error": err.Error(),
        }).Warn("No open port found in portrange, using random port")
        // TODO get a random port to use for the client
    }
}

func main() {
    defer logFile.Close()
    defer log.Info("Graytorrent stopped")

    // Single file torrent then exit
    if filename != "" {
        to, err := newTorrent(filename)
        if err != nil {
            fmt.Println("Single torrent failed:", err)
            return
        }
        to.Start()
    }

    // Initialize GUI
    // defer g.Close()

    // Send torrent stopped messages
    // Save torrent progresses to history file
}

func newTorrent(filename string) (torrent.Torrent, error) {
    to := torrent.Torrent{Path: filename}
    if err := to.Setup(); err != nil {
        log.WithFields(log.Fields{"file": filename, "error": err.Error()}).Info("Torrent setup failed")
        return torrent.Torrent{}, err
    }
    torrentList = append(torrentList, to)
    log.WithField("name", to.Info.Name).Info("Torrent added")
    return to, nil
}
