package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func GruvboxDark() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#1D2021")
	colors.Base.Foreground = lipgloss.Color("#EBDBB2")
	colors.Base.Surface = lipgloss.Color("#282828")
	colors.Base.Border = lipgloss.Color("#504945")
	colors.Base.Accent = lipgloss.Color("#8EC07C")
	colors.Base.Error = lipgloss.Color("#FB4934")
	colors.Base.Dimmed = lipgloss.Color("#A89984")
	colors.Base.Subtle = lipgloss.Color("#A89984")
	colors.Base.OnAccent = lipgloss.Color("#1D2021")
	colors.Base.Divider = lipgloss.Color("#282828")

	colors.Sidebar.Server = lipgloss.Color("#8EC07C")
	colors.Sidebar.Unread = lipgloss.Color("#83A598")
	colors.Sidebar.Notification = lipgloss.Color("#FABD2F")
	colors.Sidebar.Mention = lipgloss.Color("#FE8019")
	colors.Sidebar.Selection = lipgloss.Color("#282828")

	colors.Chat.Nickname = lipgloss.Color("#8EC07C")
	colors.Chat.InactiveNickname = lipgloss.Color("#A89984")
	colors.Chat.Mention = lipgloss.Color("#FB4934")
	colors.Chat.Separator = lipgloss.Color("#504945")
	colors.Chat.Self = lipgloss.Color("#B8BB26")
	colors.Chat.Connected = lipgloss.Color("#8EC07C")
	colors.Chat.Disconnected = lipgloss.Color("#FB4934")
	colors.Chat.UserEvents = lipgloss.Color("#A89984")
	colors.Chat.ServerMessage = lipgloss.Color("#FABD2F")

	colors.Palette.Border = lipgloss.Color("#8EC07C")
	colors.Palette.Highlight = lipgloss.Color("#8EC07C")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#FB4934"),
		lipgloss.Color("#B8BB26"),
		lipgloss.Color("#FABD2F"),
		lipgloss.Color("#83A598"),
		lipgloss.Color("#D3869B"),
		lipgloss.Color("#8EC07C"),
		lipgloss.Color("#FE8019"),
	}

	return styles.NewTheme(colors)
}
