package main

import (
	"fmt"
	"log"
	"os"
	"flag"

	gocui "github.com/jroimartin/gocui"
)

var logFile *os.File
var filename string
var gui *gocui.Gui
var err error

func init() {
    // Setup logging
    logFile, err = os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error: could not open log file info.log", err)
        os.Exit(1)
    }
    log.SetOutput(logFile)

    // Get filename argument for single-torrent execution
    flag.StringVar(&filename, "f", "", "Filename of torrent file")
    flag.Parse()
}

// Initialize GUI
func init() {
    // Stuff here from gui package
}

func main() {
    defer logFile.Close()
    defer gui.Close()
    log.Println("Graytorrent started")

    // Send torrent stopped messages
    // Save torrent progresses to history file

    log.Println("Successful exit")
}

