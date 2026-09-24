package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui"
)

// version is set by release builds with -ldflags "-X main.version=...".
var version string

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	tui := ui.New(cfg)
	p := tea.NewProgram(tui)

	ircClientManager := irc.NewClientManager(cfg.Servers)
	defer ircClientManager.DisconnectAll(quitMessage())

	tui.SetManager(ircClientManager)

	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

// quitMessage is the reason sent with QUIT when dex exits, as other clients
// do, e.g. "dexchat 1.1.0 - https://dexchat.org".
func quitMessage() string {
	if v := buildVersion(); v != "" {
		return "dexchat " + v + " - https://dexchat.org"
	}
	return "dexchat - https://dexchat.org"
}

// buildVersion returns the version set at link time, or the module version
// recorded by the Go toolchain, without build metadata such as "+dirty". It
// is empty for builds without either.
func buildVersion() string {
	v := version
	if v == "" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
	}
	v, _, _ = strings.Cut(v, "+")
	return strings.TrimPrefix(v, "v")
}
