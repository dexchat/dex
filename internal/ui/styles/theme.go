package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Colors Colors
	Styles Styles
}

type Color = color.Color

type Colors struct {
	Base struct {
		Background Color
		Foreground Color
		Surface    Color
		Border     Color
		Accent     Color
		Error      Color
		Dimmed     Color
		Subtle     Color
	}

	Sidebar struct {
		Server       Color
		Unread       Color
		Notification Color
		Mention      Color
		Selection    Color
	}

	Chat struct {
		Nickname      Color
		Mention       Color
		Separator     Color
		Self          Color
		Connected     Color
		Disconnected  Color
		UserEvents    Color
		ServerMessage Color
	}

	Palette struct {
		Border    Color
		Highlight Color
	}

	Nicknames []Color
}

type Styles struct {
	App           lipgloss.Style
	Sidebar       lipgloss.Style
	UnreadItem    lipgloss.Style
	NotifiedItem  lipgloss.Style
	MentionedItem lipgloss.Style
	ServerItem    lipgloss.Style
	ChatArea      lipgloss.Style
	InputField    lipgloss.Style
}
