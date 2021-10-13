package client

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const refreshTime = 1 * time.Second

var (
	// ErrClear for platforms where clearing the terminal is not supported
	ErrClear = errors.New("Failed to clear the terminal, platform not supported")
	clear    map[string]func() //create a map for storing clear funcs
)

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func callClear() error {
	clearFunc, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                              //if we defined a clear func for that platform:
		clearFunc()
	} else {
		return ErrClear
	}
	return nil
}

// Monitor updates a list of managed torrents on the terminal
func Monitor() error {
	for {
		if err := callClear(); err != nil {
			return err
		} else if err = List(); err != nil {
			return err
		}
		time.Sleep(refreshTime)
	}
}
