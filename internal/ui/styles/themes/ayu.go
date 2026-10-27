package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func AyuDark() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#0D1017")
	colors.Base.Foreground = lipgloss.Color("#B3B1AD")
	colors.Base.Surface = lipgloss.Color("#131721")
	colors.Base.Border = lipgloss.Color("#253340")
	colors.Base.Accent = lipgloss.Color("#59C2FF")
	colors.Base.Error = lipgloss.Color("#FF3333")
	colors.Base.Dimmed = lipgloss.Color("#949BA4")
	colors.Base.Subtle = lipgloss.Color("#75808A")
	colors.Base.OnAccent = lipgloss.Color("#0D1017")
	colors.Base.Divider = lipgloss.Color("#131721")

	colors.Sidebar.Server = lipgloss.Color("#59C2FF")
	colors.Sidebar.Unread = lipgloss.Color("#73D0FF")
	colors.Sidebar.Notification = lipgloss.Color("#FFB454")
	colors.Sidebar.Mention = lipgloss.Color("#F07178")
	colors.Sidebar.Selection = lipgloss.Color("#253340")

	colors.Chat.Nickname = lipgloss.Color("#59C2FF")
	colors.Chat.InactiveNickname = lipgloss.Color("#7D7D7D")
	colors.Chat.Mention = lipgloss.Color("#F07178")
	colors.Chat.Separator = lipgloss.Color("#253340")
	colors.Chat.Self = lipgloss.Color("#BAE67E")
	colors.Chat.Connected = lipgloss.Color("#95E6CB")
	colors.Chat.Disconnected = lipgloss.Color("#F07178")
	colors.Chat.UserEvents = lipgloss.Color("#75808A")
	colors.Chat.ServerMessage = lipgloss.Color("#E8CA20")

	colors.Palette.Border = lipgloss.Color("#59C2FF")
	colors.Palette.Highlight = lipgloss.Color("#59C2FF")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#39BAE6"),
		lipgloss.Color("#FFB454"),
		lipgloss.Color("#59C2FF"),
		lipgloss.Color("#AAD94C"),
		lipgloss.Color("#95E6CB"),
		lipgloss.Color("#F07178"),
		lipgloss.Color("#FF8F40"),
		lipgloss.Color("#D2A6FF"),
	}

	return styles.NewTheme(colors)
}
