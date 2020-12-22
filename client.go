package main

import (
    "fmt"
    // "flag"
    "os"
    "os/exec"
    "runtime"
)

var clear map[string]func()

func init() {
    clear = make(map[string]func())
    clear["linux"] = func() {
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }
    clear["windows"] = func() {
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }

    // Load settings
    // Check for existing torrents
}

func Clear() {
    value, ok := clear[runtime.GOOS]
    if ok {
        value()
    } else {
        panic("Your platform is unsupported.")
    }
}

func PrintStatus(s string) {
    fmt.Printf("--- Status --- %s\n", s)
    // for torrent in range(torrents)
    // torrent.Print()
}

func main() {
    var input, result string
    for input != "quit" {
        Clear()
        PrintStatus(result)
        fmt.Printf("-> ")
        switch fmt.Scanln(&input); input {
        case "new":
            fmt.Printf("Filename -> ")
            fmt.Scanln(&input)
            result = "[new]"
            // make Torrent(input)
        case "start":
            fmt.Printf("Start -> ")
            fmt.Scanln(&input)
            result = "[start]"
        case "stop":
            fmt.Printf("Stop -> ")
            fmt.Scanln(&input)
            result = "[stop]"
        case "remove":
            fmt.Printf("Remove -> ")
            fmt.Scanln(&input)
            result = "[remove]"
        default:
            result = "[Command " + input + " not recognized]"
        }
    }
    // Send torrent stopped messages
    // Save torrent progresses to history file
}
