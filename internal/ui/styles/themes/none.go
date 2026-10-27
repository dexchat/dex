package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func None() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.NoColor{}
	colors.Base.Foreground = lipgloss.NoColor{}
	colors.Base.Surface = lipgloss.NoColor{}
	colors.Base.Border = lipgloss.Color("8")
	colors.Base.Accent = lipgloss.Color("4")
	colors.Base.Error = lipgloss.Color("1")
	colors.Base.Dimmed = lipgloss.Color("8")
	colors.Base.Subtle = lipgloss.Color("8")
	colors.Base.OnAccent = lipgloss.Color("0")
	colors.Base.Divider = lipgloss.Color("8")

	colors.Sidebar.Server = lipgloss.Color("4")
	colors.Sidebar.Unread = lipgloss.Color("6")
	colors.Sidebar.Notification = lipgloss.Color("3")
	colors.Sidebar.Mention = lipgloss.Color("1")
	colors.Sidebar.Selection = lipgloss.NoColor{}

	colors.Chat.Nickname = lipgloss.Color("4")
	colors.Chat.InactiveNickname = lipgloss.Color("8")
	colors.Chat.Mention = lipgloss.Color("1")
	colors.Chat.Separator = lipgloss.Color("8")
	colors.Chat.Self = lipgloss.Color("2")
	colors.Chat.Connected = lipgloss.Color("2")
	colors.Chat.Disconnected = lipgloss.Color("1")
	colors.Chat.UserEvents = lipgloss.Color("8")
	colors.Chat.ServerMessage = lipgloss.Color("3")

	colors.Palette.Border = lipgloss.Color("4")
	colors.Palette.Highlight = lipgloss.Color("4")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("2"),
		lipgloss.Color("3"),
		lipgloss.Color("4"),
		lipgloss.Color("5"),
		lipgloss.Color("6"),
		lipgloss.Color("10"),
		lipgloss.Color("11"),
		lipgloss.Color("12"),
		lipgloss.Color("13"),
		lipgloss.Color("14"),
	}

	return styles.NewTheme(colors)
}
