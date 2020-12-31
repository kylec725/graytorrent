package gui

import (
    "log"
    "fmt"

    gocui "github.com/jroimartin/gocui"
)

// Setup initializes a new GUI session
func Setup() *gocui.Gui {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Fatalln("Fatal: Interface failed to load", err)
    }

    g.SetManagerFunc(layout)

    if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
        log.Panicln(err)
    }
    // if err := g.SetKeybinding("", gocui.KeyCtrlN, gocui.Mod, quit); err != nil {
    //     log.Panicln(err)
    // }


    if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
        log.Panicln(err)
    }

    return g
}

func layout(g *gocui.Gui) error {
    // Grab max values to create views
    maxX, maxY := g.Size()
    // maxX, maxY = maxX - 1, maxY - 1

    // Command view
    if v, err := g.SetView("command", 0, 0, maxX/4, maxY/4); err != nil {
        if err != gocui.ErrUnknownView {
            // log.Fatalln("Fatal: Error creating view", err)
            return err
        }
        fmt.Fprintln(v, "Command view")
    }

    // Filter view
    if v, err := g.SetView("filter", 0, (maxY/4) + 1, maxX/4, maxY - 1); err != nil {
        if err != gocui.ErrUnknownView {
            // log.Fatalln("Fatal: Error creating view", err)
            return err
        }
        fmt.Fprintln(v, "Filter view")
    }

    // Status view of current torrents
    if v, err := g.SetView("status", maxX/4 + 1, 0, maxX - 1, (maxY/3) * 2); err != nil {
        if err != gocui.ErrUnknownView {
            // log.Fatalln("Fatal: Error creating view", err)
            return err
        }
        fmt.Fprintln(v, "Status view")
    }

    // Detail view of a specific torrent
    if v, err := g.SetView("detail", maxX/4 + 1, (2*(maxY/3)) + 1, maxX - 1, maxY - 1); err != nil {
        if err != gocui.ErrUnknownView {
            // log.Fatalln("Fatal: Error creating view", err)
            return err
        }
        fmt.Fprintln(v, "Detail view")
    }

    return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}
