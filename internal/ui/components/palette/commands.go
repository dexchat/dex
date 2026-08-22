package palette

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
)

type action struct {
	Name        string
	Description string
	Keybinding  key.Binding
	Handler     func() tea.Msg // Function to execute when command is selected
}

type OpenChannelPickerMsg struct{}

type LastBufferMsg struct{}

type OpenHelpMsg struct{}

type ChannelSelectionMsg struct {
	Server  string
	Channel string
}

// matches checks if the command matches the given input query (case-insensitive)
func (a action) matches(query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(a.Name), query) ||
		strings.Contains(strings.ToLower(a.Description), query)
}

// defaultActions does not necessarily display all keybindings available, only
// the ones I believe should be shown in a command palette for fast selection
func defaultActions() []action {
	kb := keybindings.DefaultKeyMap()
	return []action{
		{
			Name:        kb.GoToChannel.Help().Desc,
			Description: "open a connected channel",
			Keybinding:  kb.GoToChannel,
			Handler:     func() tea.Msg { return OpenChannelPickerMsg{} },
		},
		{
			Name:        kb.LastBuffer.Help().Desc,
			Description: "switch to the last selected channel",
			Keybinding:  kb.LastBuffer,
			Handler:     func() tea.Msg { return LastBufferMsg{} },
		},
		{
			Name:        kb.EditInEditor.Help().Desc,
			Description: "edit the message in your default editor",
			Keybinding:  kb.EditInEditor,
			Handler:     func() tea.Msg { return nil },
		},
		{
			Name:        "help",
			Description: "show help and keybindings",
			Keybinding:  key.NewBinding(),
			Handler:     func() tea.Msg { return OpenHelpMsg{} },
		},
		{
			Name:        kb.Quit.Help().Desc,
			Description: "exit dexchat",
			Keybinding:  kb.Quit,
			Handler:     func() tea.Msg { return tea.Quit() },
		},
	}
}
