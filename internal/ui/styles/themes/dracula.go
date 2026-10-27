package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func Dracula() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#282A36")
	colors.Base.Foreground = lipgloss.Color("#F8F8F2")
	colors.Base.Surface = lipgloss.Color("#343746")
	colors.Base.Border = lipgloss.Color("#6272A4")
	colors.Base.Accent = lipgloss.Color("#BD93F9")
	colors.Base.Error = lipgloss.Color("#FF5555")
	colors.Base.Dimmed = lipgloss.Color("#B8B8C2")
	colors.Base.Subtle = lipgloss.Color("#B8B8C2")
	colors.Base.OnAccent = lipgloss.Color("#282A36")
	colors.Base.Divider = lipgloss.Color("#343746")

	colors.Sidebar.Server = lipgloss.Color("#8BE9FD")
	colors.Sidebar.Unread = lipgloss.Color("#D6ACFF")
	colors.Sidebar.Notification = lipgloss.Color("#F1FA8C")
	colors.Sidebar.Mention = lipgloss.Color("#FF92DF")
	colors.Sidebar.Selection = lipgloss.Color("#44475A")

	colors.Chat.Nickname = lipgloss.Color("#8BE9FD")
	colors.Chat.InactiveNickname = lipgloss.Color("#B8B8C2")
	colors.Chat.Mention = lipgloss.Color("#FF79C6")
	colors.Chat.Separator = lipgloss.Color("#6272A4")
	colors.Chat.Self = lipgloss.Color("#50FA7B")
	colors.Chat.Connected = lipgloss.Color("#50FA7B")
	colors.Chat.Disconnected = lipgloss.Color("#FF5555")
	colors.Chat.UserEvents = lipgloss.Color("#B8B8C2")
	colors.Chat.ServerMessage = lipgloss.Color("#FFB86C")

	colors.Palette.Border = lipgloss.Color("#BD93F9")
	colors.Palette.Highlight = lipgloss.Color("#BD93F9")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#FF5555"),
		lipgloss.Color("#FFB86C"),
		lipgloss.Color("#F1FA8C"),
		lipgloss.Color("#50FA7B"),
		lipgloss.Color("#8BE9FD"),
		lipgloss.Color("#BD93F9"),
		lipgloss.Color("#FF79C6"),
	}

	return styles.NewTheme(colors)
}
