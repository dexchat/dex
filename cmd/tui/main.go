package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	tui := ui.New(cfg)
	p := tea.NewProgram(tui)

	ircClientManager := irc.NewClientManager(cfg.Servers, p)
	defer ircClientManager.DisconnectAll()

	tui.SetManager(ircClientManager)

	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
