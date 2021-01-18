package main

import (
	"fmt"
	"log"
	"os"

    // "github.com/kylec725/graytorrent/gui"
    "github.com/kylec725/graytorrent/connect"
    flag "github.com/spf13/pflag"
    viper "github.com/spf13/viper"
	// gocui "github.com/jroimartin/gocui"
)

var (
    logFile *os.File
    port uint16
    err error
    // g *gocui.Gui

    filename string
)

func init() {
    // Setup logging
    logFile, err = os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Fatal: could not open log file info.log", err)
        os.Exit(1)
    }
    log.SetOutput(logFile)
    log.Println("Graytorrent started")

    flag.StringVarP(&filename, "file", "f", "", "Filename of torrent file")
    flag.Parse()
}

// Initialize GUI
func init() {
    // g = gui.Setup()
}

// Initialize config
func init() {
    viper.SetDefault("torrent.path", ".")
    viper.SetDefault("torrent.autoseed", true)
    viper.SetDefault("network.portrange", [2]int{ 6881, 6889 })
    viper.SetDefault("network.connections.globalMax", 300)
    viper.SetDefault("network.connections.torrentMax", 30)

    viper.SetConfigName("config")
    viper.SetConfigType("toml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("~/.config/graytorrent")
    viper.AddConfigPath("/etc/graytorrent")

    if err = viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found, create default config
            viper.SafeWriteConfig()
        } else {
            // Some other error was found
            log.Panic("Fatal error reading in config file:", err)
        }
    }

    port, err = connect.OpenPort(viper.GetIntSlice("network.portrange"))
    if err != nil {
        log.Fatalln("Open port could not be obtained for the client:", err)
    }

    // Check for live changes of the config file
    viper.WatchConfig()
}

func main() {
    defer logFile.Close()
    // defer g.Close()

    // Send torrent stopped messages
    // Save torrent progresses to history file

    log.Println("Successful exit")
}

