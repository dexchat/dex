package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vaaleyard/dex/internal/ui"
	"os"
)

func main() {
	p := tea.NewProgram(ui.New())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
