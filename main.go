package main

import (
	"fmt"
	// "flag" // want to implement a single download -f mode and accompanying -v verbose
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

// Clear terminal output
func Clear() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		panic("Your platform is unsupported.")
	}
}

// PrintStatus returns current torrents and their status
func PrintStatus(s string) {
	fmt.Println("--- Commands [new] [start] [stop] [remove] [quit]")
	fmt.Printf("--- Status %s\n", s)
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
