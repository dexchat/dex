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

func GruvboxLight() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#F9F5D7")
	colors.Base.Foreground = lipgloss.Color("#3C3836")
	colors.Base.Surface = lipgloss.Color("#EBDBB2")
	colors.Base.Border = lipgloss.Color("#D5C4A1")
	colors.Base.Accent = lipgloss.Color("#38694B")
	colors.Base.Error = lipgloss.Color("#9D0006")
	colors.Base.Dimmed = lipgloss.Color("#685D54")
	colors.Base.Subtle = lipgloss.Color("#685D54")
	colors.Base.OnAccent = lipgloss.Color("#F9F5D7")
	colors.Base.Divider = lipgloss.Color("#EBDBB2")

	colors.Sidebar.Server = lipgloss.Color("#38694B")
	colors.Sidebar.Unread = lipgloss.Color("#076678")
	colors.Sidebar.Notification = lipgloss.Color("#84560F")
	colors.Sidebar.Mention = lipgloss.Color("#AB3903")
	colors.Sidebar.Selection = lipgloss.Color("#EBDBB2")

	colors.Chat.Nickname = lipgloss.Color("#38694B")
	colors.Chat.InactiveNickname = lipgloss.Color("#685D54")
	colors.Chat.Mention = lipgloss.Color("#9D0006")
	colors.Chat.Separator = lipgloss.Color("#D5C4A1")
	colors.Chat.Self = lipgloss.Color("#76710E")
	colors.Chat.Connected = lipgloss.Color("#38694B")
	colors.Chat.Disconnected = lipgloss.Color("#9D0006")
	colors.Chat.UserEvents = lipgloss.Color("#685D54")
	colors.Chat.ServerMessage = lipgloss.Color("#84560F")

	colors.Palette.Border = lipgloss.Color("#38694B")
	colors.Palette.Highlight = lipgloss.Color("#38694B")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#9D0006"),
		lipgloss.Color("#76710E"),
		lipgloss.Color("#84560F"),
		lipgloss.Color("#076678"),
		lipgloss.Color("#8F3F71"),
		lipgloss.Color("#38694B"),
		lipgloss.Color("#AB3903"),
	}

	return styles.NewTheme(colors)
}
