package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui"
)

func main() {
	// for debugging IRC events only
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(cfg), tea.WithAltScreen(), tea.WithMouseCellMotion())
	client := irc.NewClient("libera", cfg.Servers["libera"], p)

	go func() {
		if err := client.Connect(); err != nil {
			fmt.Printf("Failed to connect to %s: %v\n", "libera", err)
		}
	}()

	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
