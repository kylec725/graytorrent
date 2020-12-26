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
    // gui, err = gocui.NewGui(gocui.OutputNormal)
    // if err != nil {
    //     log.Fatalln("Fatal: Interface failed to load", err)
    // }
}

func main() {
    defer logFile.Close()
    defer gui.Close()
    log.Println("Graytorrent started")

    // if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
    //     log.Fatal("Fatal: Interface crashed.", err)
    // }

    // Send torrent stopped messages
    // Save torrent progresses to history file

    log.Println("Successful exit")
}

func layout(gui *gocui.Gui) error {
    // Grab max values to create views
    maxX, maxY := gui.Size()
    if _, err := gui.SetView("xd", 10, 10, maxX-10, maxY-10); err != gocui.ErrUnknownView {
        log.Fatalln("Fatal: Error creating view", err)
    }

    return nil
}
