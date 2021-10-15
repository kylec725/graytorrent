package ui

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

var (
	serverAddr string
)

// Run launches a TUI session
func Run() {
	serverAddr = "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Could not start terminal UI:", err)
		os.Exit(1)
	}
}
