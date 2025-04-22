package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vaaleyard/dex/internal/ui"
	"os"
)

func main() {
	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
