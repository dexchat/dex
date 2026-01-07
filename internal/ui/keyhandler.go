package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
)

func (m *model) handleKeybindings(msg tea.KeyMsg) tea.Msg {
	kb := keybindings.DefaultKeyMap()

	if key.Matches(msg, kb.TogglePalette) {
		m.palette.Toggle()
		return nil
	}

	// handling palette keys here ended up needing flags to conditionally update the palette
	// and I couldn't find a good/working way to do it
	if m.palette.IsVisible() {
		return msg
	}

	switch {
	case key.Matches(msg, kb.MoveUp):
		m.channels.MoveUp()
		return nil
	case key.Matches(msg, kb.MoveDown):
		m.channels.MoveDown()
		return nil
	}

	return msg
}
