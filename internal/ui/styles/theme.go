package styles

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Colors struct {
		Background   lipgloss.Color
		Text         lipgloss.Color
		Accent       lipgloss.Color
		SidebarBg    lipgloss.Color
		InputBg      lipgloss.Color
		BorderColor  lipgloss.Color
		SelfMsg      lipgloss.Color
		SystemMsg    lipgloss.Color
		ErrorMsg     lipgloss.Color
		Timestamp    lipgloss.Color
		MentionColor lipgloss.Color
		UnreadColor  lipgloss.Color
		StatusBg     lipgloss.Color
		StatusText   lipgloss.Color
		Usernames    []lipgloss.Color
	}

	Styles struct {
		App           lipgloss.Style
		Sidebar       lipgloss.Style
		SelectedItem  lipgloss.Style
		UnreadItem    lipgloss.Style
		ChatArea      lipgloss.Style
		OwnMessage    lipgloss.Style
		SystemMessage lipgloss.Style
		ErrorMessage  lipgloss.Style
		Timestamp     lipgloss.Style
		Mention       lipgloss.Style
		InputField    lipgloss.Style
		StatusLine    lipgloss.Style
		Usernames     lipgloss.Style
	}
}
