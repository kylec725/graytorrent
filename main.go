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
    torrentList []torrent.Torrent
    listener net.Listener
)

func init() {
    flag.StringVarP(&filename, "file", "f", "", "Filename of torrent file")
    flag.Parse()

    setupLog()
    log.Info("Graytorrent started")

    setupViper()
    viper.WatchConfig()

    listener = setupListen()
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
        fmt.Println("Torrent done:", to.Info.Name)
        return
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

func peerListen(listener net.Listener) {
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.WithField("error", err).Debug("Error with incoming peer connection")
            continue
        }
        fmt.Println("New connection:", conn.LocalAddr().String())
    }
}
