package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/components/palette"
)

func (m *Model) handleKeybindings(msg tea.KeyPressMsg) tea.Msg {
	if m.help.IsVisible() {
		return msg
	}

	kb := keybindings.DefaultKeyMap()

	if key.Matches(msg, kb.TogglePalette) {
		m.palette.Toggle()
		return nil
	}
	if key.Matches(msg, kb.GoToChannel) {
		m.showChannelPicker()
		return nil
	}
	if key.Matches(msg, kb.LastBuffer) {
		return palette.LastBufferMsg{}
	}

	// Handling palette keys here ended up needing flags to conditionally update the palette
	// and I couldn't find a good/working way to do it
	if m.palette.IsVisible() {
		return msg
	}

	switch {
	case key.Matches(msg, kb.MoveUp):
		m.channels = m.channels.MoveUp()
		return m.channels.Selected()
	case key.Matches(msg, kb.MoveDown):
		m.channels = m.channels.MoveDown()
		return m.channels.Selected()
	}

	return msg
}
