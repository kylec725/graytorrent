package gui

import (
    "log"

    gocui "github.com/jroimartin/gocui"
)

// Initialize a new GUI
func setup() {
    gui, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Fatalln("Fatal: Interface failed to load", err)
    }

    // Grab max values to create views
    maxX, maxY := gui.Size()
    if _, err := gui.SetView("xd", 10, 10, maxX-10, maxY-10); err != gocui.ErrUnknownView {
        log.Fatalln("Fatal: Error creating view", err)
    }
}
