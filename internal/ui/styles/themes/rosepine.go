package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func RosePine() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#191724")
	colors.Base.Foreground = lipgloss.Color("#E0DEF4")
	colors.Base.Surface = lipgloss.Color("#26233A")
	colors.Base.Border = lipgloss.Color("#403D52")
	colors.Base.Accent = lipgloss.Color("#C4A7E7")
	colors.Base.Error = lipgloss.Color("#EB6F92")
	colors.Base.Dimmed = lipgloss.Color("#908CAA")
	colors.Base.Subtle = lipgloss.Color("#908CAA")
	colors.Base.OnAccent = lipgloss.Color("#191724")
	colors.Base.Divider = lipgloss.Color("#26233A")

	colors.Sidebar.Server = lipgloss.Color("#C4A7E7")
	colors.Sidebar.Unread = lipgloss.Color("#F6C177")
	colors.Sidebar.Notification = lipgloss.Color("#F6C177")
	colors.Sidebar.Mention = lipgloss.Color("#EB6F92")
	colors.Sidebar.Selection = lipgloss.Color("#26233A")

	colors.Chat.Nickname = lipgloss.Color("#EBBCBA")
	colors.Chat.InactiveNickname = lipgloss.Color("#858585")
	colors.Chat.Mention = lipgloss.Color("#EB6F92")
	colors.Chat.Separator = lipgloss.Color("#403D52")
	colors.Chat.Self = lipgloss.Color("#EBBCBA")
	colors.Chat.Connected = lipgloss.Color("#9CCFD8")
	colors.Chat.Disconnected = lipgloss.Color("#EB6F92")
	colors.Chat.UserEvents = lipgloss.Color("#908CAA")
	colors.Chat.ServerMessage = lipgloss.Color("#F6C177")

	colors.Palette.Border = lipgloss.Color("#C4A7E7")
	colors.Palette.Highlight = lipgloss.Color("#C4A7E7")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#EBBCBA"),
		lipgloss.Color("#F6C177"),
		lipgloss.Color("#9CCFD8"),
		lipgloss.Color("#C4A7E7"),
		lipgloss.Color("#EB6F92"),
		lipgloss.Color("#E0DEF4"),
	}

	return styles.NewTheme(colors)
}

func RosePineDawn() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#FAF4ED")
	colors.Base.Foreground = lipgloss.Color("#575279")
	colors.Base.Surface = lipgloss.Color("#F2E9E1")
	colors.Base.Border = lipgloss.Color("#DFDAD9")
	colors.Base.Accent = lipgloss.Color("#765E92")
	colors.Base.Error = lipgloss.Color("#A14D65")
	colors.Base.Dimmed = lipgloss.Color("#696581")
	colors.Base.Subtle = lipgloss.Color("#696581")
	colors.Base.OnAccent = lipgloss.Color("#FAF4ED")
	colors.Base.Divider = lipgloss.Color("#F2E9E1")

	colors.Sidebar.Server = lipgloss.Color("#765E92")
	colors.Sidebar.Unread = lipgloss.Color("#945C0F")
	colors.Sidebar.Notification = lipgloss.Color("#945C0F")
	colors.Sidebar.Mention = lipgloss.Color("#A14D65")
	colors.Sidebar.Selection = lipgloss.Color("#F2E9E1")

	colors.Chat.Nickname = lipgloss.Color("#935956")
	colors.Chat.InactiveNickname = lipgloss.Color("#686868")
	colors.Chat.Mention = lipgloss.Color("#A14D65")
	colors.Chat.Separator = lipgloss.Color("#DFDAD9")
	colors.Chat.Self = lipgloss.Color("#935956")
	colors.Chat.Connected = lipgloss.Color("#416F77")
	colors.Chat.Disconnected = lipgloss.Color("#A14D65")
	colors.Chat.UserEvents = lipgloss.Color("#696581")
	colors.Chat.ServerMessage = lipgloss.Color("#945C0F")

	colors.Palette.Border = lipgloss.Color("#765E92")
	colors.Palette.Highlight = lipgloss.Color("#765E92")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#935956"),
		lipgloss.Color("#945C0F"),
		lipgloss.Color("#416F77"),
		lipgloss.Color("#765E92"),
		lipgloss.Color("#A14D65"),
		lipgloss.Color("#575279"),
	}

	return styles.NewTheme(colors)
}
