package main

import (
	"fmt"
	"os"

	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/irc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	client := irc.NewClient("libera", cfg.Servers["libera"])
	if err := client.Connect(); err != nil {
		fmt.Printf("Failed to connect to %s: %v\n", "libera", err)
	}

	//p := tea.NewProgram(ui.New(cfg), tea.WithAltScreen(), tea.WithMouseCellMotion())
	//if _, err := p.Run(); err != nil {
	//	_, _ = fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
	//	os.Exit(1)
	//}
}
