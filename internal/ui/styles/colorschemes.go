package styles

import "github.com/charmbracelet/lipgloss"

func AyuDarkTheme() Theme {
	theme := Theme{}

	theme.Colors.Base.Background = "#0D1017"
	theme.Colors.Base.Foreground = "#B3B1AD"
	theme.Colors.Base.Surface = "#131721"
	theme.Colors.Base.Border = "#253340"
	theme.Colors.Base.Accent = "#59C2FF"
	theme.Colors.Base.Error = "#FF3333"
	theme.Colors.Base.Dimmed = "#626A73"
	theme.Colors.Base.Subtle = "#253340"

	theme.Colors.Sidebar.Server = "#59C2FF"
	theme.Colors.Sidebar.Unread = "#73D0FF"
	theme.Colors.Sidebar.Mention = "#F07178"
	theme.Colors.Sidebar.Selection = "#253340"

	theme.Colors.Chat.Nickname = "#59C2FF"
	theme.Colors.Chat.Separator = "#253340"
	theme.Colors.Chat.Self = "#BAE67E"
	theme.Colors.Chat.Connected = "#95E6CB"
	theme.Colors.Chat.Disconnected = "#F07178"
	theme.Colors.Chat.UserEvents = "#6E6012"
	theme.Colors.Chat.ServerMessage = "#E8CA20"

	theme.Colors.Palette.Border = "#59C2FF"
	theme.Colors.Palette.Highlight = "#59C2FF"

	theme.Colors.Nicknames = []lipgloss.Color{
		"#39BAE6",
		"#FFB454",
		"#59C2FF",
		"#AAD94C",
		"#95E6CB",
		"#F07178",
		"#FF8F40",
		"#D2A6FF",
	}

	theme.Styles.App = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Base.Foreground)

	theme.Styles.Sidebar = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Base.Foreground).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Colors.Base.Background).
		BorderBackground(theme.Colors.Base.Background)

	theme.Styles.UnreadItem = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Sidebar.Unread)

	theme.Styles.MentionedItem = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Sidebar.Mention).
		Bold(true)

	theme.Styles.ServerItem = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Sidebar.Server).
		Bold(true)

	theme.Styles.ChatArea = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Base.Foreground).
		BorderForeground(theme.Colors.Base.Background).
		BorderBackground(theme.Colors.Base.Background).
		Border(lipgloss.NormalBorder())

	theme.Styles.InputField = lipgloss.NewStyle().
		Background(theme.Colors.Base.Surface).
		Foreground(theme.Colors.Base.Foreground)

	return theme
}
