package main

import (
	"fmt"
	"log"
	"os"
	"flag"

    // "github.com/kylec725/graytorrent/gui"
    viper "github.com/spf13/viper"
	// gocui "github.com/jroimartin/gocui"
)

var logFile *os.File
var filename string
// var g *gocui.Gui
var err error

func init() {
    // Setup logging
    logFile, err = os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error: could not open log file info.log", err)
        os.Exit(1)
    }
    log.SetOutput(logFile)
    log.Println("Graytorrent started")

    // Get filename argument for single-torrent execution
    flag.StringVar(&filename, "f", "", "Filename of torrent file")
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
    viper.SetDefault("network.port", 6881)
    viper.SetDefault("network.connections.globalMax", 300)
    viper.SetDefault("network.connections.torrentMax", 30)

    viper.SetConfigName("config")
    viper.SetConfigType("toml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("~/.config/graytorrent")
    viper.AddConfigPath("/etc/graytorrent")

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found, create default config
            viper.SafeWriteConfig()
        } else {
            log.Println("Error reading in config file:", err)
        }
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

