package styles

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Colors Colors
	Styles Styles
}

type Colors struct {
	Base struct {
		Background lipgloss.Color
		Foreground lipgloss.Color
		Surface    lipgloss.Color
		Border     lipgloss.Color
		Accent     lipgloss.Color
		Error      lipgloss.Color
		Dimmed     lipgloss.Color
		Subtle     lipgloss.Color
	}

	Sidebar struct {
		Server    lipgloss.Color
		Unread    lipgloss.Color
		Mention   lipgloss.Color
		Selection lipgloss.Color
	}

	Chat struct {
		Nickname      lipgloss.Color
		Separator     lipgloss.Color
		Self          lipgloss.Color
		Connected     lipgloss.Color
		Disconnected  lipgloss.Color
		UserEvents    lipgloss.Color
		ServerMessage lipgloss.Color
	}

	Palette struct {
		Border    lipgloss.Color
		Highlight lipgloss.Color
	}

	Nicknames []lipgloss.Color
}

type Styles struct {
	App           lipgloss.Style
	Sidebar       lipgloss.Style
	UnreadItem    lipgloss.Style
	MentionedItem lipgloss.Style
	ServerItem    lipgloss.Style
	ChatArea      lipgloss.Style
	InputField    lipgloss.Style
}
