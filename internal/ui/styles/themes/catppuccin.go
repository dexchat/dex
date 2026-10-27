package themes

import (
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func CatppuccinLatte() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#EFF1F5")
	colors.Base.Foreground = lipgloss.Color("#4C4F69")
	colors.Base.Surface = lipgloss.Color("#DCE0E8")
	colors.Base.Border = lipgloss.Color("#CCD0DA")
	colors.Base.Accent = lipgloss.Color("#7E29EE")
	colors.Base.Error = lipgloss.Color("#C30E35")
	colors.Base.Dimmed = lipgloss.Color("#5F6173")
	colors.Base.Subtle = lipgloss.Color("#5F6173")
	colors.Base.OnAccent = lipgloss.Color("#EFF1F5")
	colors.Base.Divider = lipgloss.Color("#DCE0E8")

	colors.Sidebar.Server = lipgloss.Color("#7E29EE")
	colors.Sidebar.Unread = lipgloss.Color("#0A54E7")
	colors.Sidebar.Notification = lipgloss.Color("#895712")
	colors.Sidebar.Mention = lipgloss.Color("#C30E35")
	colors.Sidebar.Selection = lipgloss.Color("#DCE0E8")

	colors.Chat.Nickname = lipgloss.Color("#3C59FC")
	colors.Chat.InactiveNickname = lipgloss.Color("#5F6173")
	colors.Chat.Mention = lipgloss.Color("#C30E35")
	colors.Chat.Separator = lipgloss.Color("#CCD0DA")
	colors.Chat.Self = lipgloss.Color("#317B21")
	colors.Chat.Connected = lipgloss.Color("#317B21")
	colors.Chat.Disconnected = lipgloss.Color("#C30E35")
	colors.Chat.UserEvents = lipgloss.Color("#5F6173")
	colors.Chat.ServerMessage = lipgloss.Color("#BC4501")

	colors.Palette.Border = lipgloss.Color("#7E29EE")
	colors.Palette.Highlight = lipgloss.Color("#7E29EE")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#CA3333"),
		lipgloss.Color("#C51E99"),
		lipgloss.Color("#7E29EE"),
		lipgloss.Color("#BC4501"),
		lipgloss.Color("#895712"),
		lipgloss.Color("#317B21"),
		lipgloss.Color("#13787E"),
		lipgloss.Color("#0374A1"),
		lipgloss.Color("#0A54E7"),
		lipgloss.Color("#3C59FC"),
	}

	return styles.NewTheme(colors)
}

func CatppuccinMocha() styles.Theme {
	var colors styles.Colors

	colors.Base.Background = lipgloss.Color("#1E1E2E")
	colors.Base.Foreground = lipgloss.Color("#CDD6F4")
	colors.Base.Surface = lipgloss.Color("#313244")
	colors.Base.Border = lipgloss.Color("#45475A")
	colors.Base.Accent = lipgloss.Color("#CBA6F7")
	colors.Base.Error = lipgloss.Color("#F38BA8")
	colors.Base.Dimmed = lipgloss.Color("#969CB4")
	colors.Base.Subtle = lipgloss.Color("#969CB4")
	colors.Base.OnAccent = lipgloss.Color("#1E1E2E")
	colors.Base.Divider = lipgloss.Color("#313244")

	colors.Sidebar.Server = lipgloss.Color("#CBA6F7")
	colors.Sidebar.Unread = lipgloss.Color("#89B4FA")
	colors.Sidebar.Notification = lipgloss.Color("#F9E2AF")
	colors.Sidebar.Mention = lipgloss.Color("#F38BA8")
	colors.Sidebar.Selection = lipgloss.Color("#313244")

	colors.Chat.Nickname = lipgloss.Color("#B4BEFE")
	colors.Chat.InactiveNickname = lipgloss.Color("#969CB4")
	colors.Chat.Mention = lipgloss.Color("#F38BA8")
	colors.Chat.Separator = lipgloss.Color("#45475A")
	colors.Chat.Self = lipgloss.Color("#A6E3A1")
	colors.Chat.Connected = lipgloss.Color("#A6E3A1")
	colors.Chat.Disconnected = lipgloss.Color("#F38BA8")
	colors.Chat.UserEvents = lipgloss.Color("#969CB4")
	colors.Chat.ServerMessage = lipgloss.Color("#FAB387")

	colors.Palette.Border = lipgloss.Color("#CBA6F7")
	colors.Palette.Highlight = lipgloss.Color("#CBA6F7")

	colors.Nicknames = []styles.Color{
		lipgloss.Color("#F2CDCD"),
		lipgloss.Color("#F5C2E7"),
		lipgloss.Color("#CBA6F7"),
		lipgloss.Color("#FAB387"),
		lipgloss.Color("#F9E2AF"),
		lipgloss.Color("#A6E3A1"),
		lipgloss.Color("#94E2D5"),
		lipgloss.Color("#89DCEB"),
		lipgloss.Color("#89B4FA"),
		lipgloss.Color("#B4BEFE"),
	}

	return styles.NewTheme(colors)
}
