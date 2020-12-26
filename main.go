package main

import (
	"fmt"
	"log"
	"os"
	// "flag" // want to implement a single download -f mode and accompanying -v verbose

	gocui "github.com/jroimartin/gocui"
)

func main() {
    // Setup logging
    logFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error opening log file", err)
        os.Exit(1)
    }
    defer logFile.Close()
    log.SetOutput(logFile)
    log.Println("Graytorrent started")

    gui, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Fatalln("Fatal: Interface failed to load", err)
    }
    defer gui.Close()

    // Grab max values to create views
    maxX, maxY := gui.Size()
    if _, err := gui.SetView("xd", 10, 10, maxX-10, maxY-10); err != gocui.ErrUnknownView {
        log.Fatalln("Fatal: Error creating view", err)
    }

    // if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
    //     log.Fatal("Fatal: Interface crashed.", err)
    // }

    // Send torrent stopped messages
    // Save torrent progresses to history file

    log.Println("Successful exit")
}
