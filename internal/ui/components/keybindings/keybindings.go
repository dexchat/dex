package keybindings

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	GoToChannel   key.Binding
	MoveDown      key.Binding
	MoveUp        key.Binding
	TogglePalette key.Binding
	LastBuffer    key.Binding
	EditInEditor  key.Binding
	Quit          key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToChannel: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "go to channel"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "move cursor down"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "move cursor up"),
		),
		TogglePalette: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "toggle the command palette"),
		),
		LastBuffer: key.NewBinding(
			key.WithKeys("ctrl+6", "ctrl+^"),
			key.WithHelp("ctrl+6", "last channel"),
		),
		EditInEditor: key.NewBinding(
			key.WithKeys("", ""),
			key.WithHelp("ctrl+x+e", "edit in editor"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
