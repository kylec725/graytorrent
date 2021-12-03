package ui

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kylec725/graytorrent/internal/config"
)

var (
	serverAddr string
)

// Run launches a TUI session
func Run() {
	cfg := config.GetConfig()
	serverAddr = cfg.Network.ServerAddress + ":" + strconv.Itoa(cfg.Network.ServerPort)

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Could not start terminal UI:", err)
		os.Exit(1)
	}
}
