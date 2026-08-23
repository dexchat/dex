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
	ScrollUp      key.Binding
	ScrollDown    key.Binding
	Autocomplete  key.Binding
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
		ScrollUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "scroll chat up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "scroll chat down"),
		),
		Autocomplete: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "autocomplete nickname"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
