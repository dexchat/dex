package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func SolarizedLight() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#FDF6E3")
	colors.Base.Foreground = lipgloss.Color("#4E646B")
	colors.Base.Surface = lipgloss.Color("#EEE8D5")
	colors.Base.Border = lipgloss.Color("#93A1A1")
	colors.Base.Accent = lipgloss.Color("#1B6EA3")
	colors.Base.Error = lipgloss.Color("#C23130")
	colors.Base.Dimmed = lipgloss.Color("#556B72")
	colors.Base.Subtle = lipgloss.Color("#556B72")
	colors.Base.OnAccent = lipgloss.Color("#FDF6E3")
	colors.Base.Divider = lipgloss.Color("#EEE8D5")

	colors.Sidebar.Server = lipgloss.Color("#1A7473")
	colors.Sidebar.Unread = lipgloss.Color("#1B6EA3")
	colors.Sidebar.Notification = lipgloss.Color("#766813")
	colors.Sidebar.Mention = lipgloss.Color("#B1471A")
	colors.Sidebar.Selection = lipgloss.Color("#EEE8D5")

	colors.Chat.Nickname = lipgloss.Color("#1B6EA3")
	colors.Chat.InactiveNickname = lipgloss.Color("#556B72")
	colors.Chat.Mention = lipgloss.Color("#B53477")
	colors.Chat.Separator = lipgloss.Color("#93A1A1")
	colors.Chat.Self = lipgloss.Color("#557113")
	colors.Chat.Connected = lipgloss.Color("#1A7473")
	colors.Chat.Disconnected = lipgloss.Color("#C23130")
	colors.Chat.UserEvents = lipgloss.Color("#556B72")
	colors.Chat.ServerMessage = lipgloss.Color("#B1471A")

	colors.Palette.Border = lipgloss.Color("#1B6EA3")
	colors.Palette.Highlight = lipgloss.Color("#1B6EA3")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#766813"),
		lipgloss.Color("#B1471A"),
		lipgloss.Color("#C23130"),
		lipgloss.Color("#B53477"),
		lipgloss.Color("#5764A9"),
		lipgloss.Color("#1B6EA3"),
		lipgloss.Color("#1A7473"),
		lipgloss.Color("#557113"),
	}

	return styles.NewTheme(colors)
}

func SolarizedDark() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#002B36")
	colors.Base.Foreground = lipgloss.Color("#8D9D9E")
	colors.Base.Surface = lipgloss.Color("#073642")
	colors.Base.Border = lipgloss.Color("#586E75")
	colors.Base.Accent = lipgloss.Color("#4AA1DE")
	colors.Base.Error = lipgloss.Color("#E56765")
	colors.Base.Dimmed = lipgloss.Color("#879EA5")
	colors.Base.Subtle = lipgloss.Color("#879EA5")
	colors.Base.OnAccent = lipgloss.Color("#002B36")
	colors.Base.Divider = lipgloss.Color("#073642")

	colors.Sidebar.Server = lipgloss.Color("#2DABA1")
	colors.Sidebar.Unread = lipgloss.Color("#4AA1DE")
	colors.Sidebar.Notification = lipgloss.Color("#C29300")
	colors.Sidebar.Mention = lipgloss.Color("#EC7A4B")
	colors.Sidebar.Selection = lipgloss.Color("#073642")

	colors.Chat.Nickname = lipgloss.Color("#4AA1DE")
	colors.Chat.InactiveNickname = lipgloss.Color("#879EA5")
	colors.Chat.Mention = lipgloss.Color("#DD659F")
	colors.Chat.Separator = lipgloss.Color("#586E75")
	colors.Chat.Self = lipgloss.Color("#859900")
	colors.Chat.Connected = lipgloss.Color("#2DABA1")
	colors.Chat.Disconnected = lipgloss.Color("#E56765")
	colors.Chat.UserEvents = lipgloss.Color("#879EA5")
	colors.Chat.ServerMessage = lipgloss.Color("#EC7A4B")

	colors.Palette.Border = lipgloss.Color("#4AA1DE")
	colors.Palette.Highlight = lipgloss.Color("#4AA1DE")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#C29300"),
		lipgloss.Color("#EC7A4B"),
		lipgloss.Color("#E56765"),
		lipgloss.Color("#DD659F"),
		lipgloss.Color("#8589CE"),
		lipgloss.Color("#4AA1DE"),
		lipgloss.Color("#2DABA1"),
		lipgloss.Color("#859900"),
	}

	return styles.NewTheme(colors)
}
