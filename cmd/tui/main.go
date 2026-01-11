package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	for name, server := range cfg.Servers {
		client := irc.NewClient(name, server, p)

		go func(name string, client *irc.Client) {
			if err := client.Connect(); err != nil {
				p.Send(irc.BufferNewMessageMsg{
					Server:  name,
					Channel: "",
					Time:    time.Now().Format("15:04"),
					From:    name,
					Text:    err.Error(),
				})
			}
		}(name, client)
	}

	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
