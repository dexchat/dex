package styles

import "charm.land/lipgloss/v2"

func AyuDarkTheme() Theme {
	theme := Theme{}

	theme.Colors.Base.Background = lipgloss.Color("#0D1017")
	theme.Colors.Base.Foreground = lipgloss.Color("#B3B1AD")
	theme.Colors.Base.Surface = lipgloss.Color("#131721")
	theme.Colors.Base.Border = lipgloss.Color("#253340")
	theme.Colors.Base.Accent = lipgloss.Color("#59C2FF")
	theme.Colors.Base.Error = lipgloss.Color("#FF3333")
	theme.Colors.Base.Dimmed = lipgloss.Color("#626A73")
	theme.Colors.Base.Subtle = lipgloss.Color("#253340")

	theme.Colors.Sidebar.Server = lipgloss.Color("#59C2FF")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#73D0FF")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#FFB454")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#F07178")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#253340")

	theme.Colors.Chat.Nickname = lipgloss.Color("#59C2FF")
	theme.Colors.Chat.Separator = lipgloss.Color("#253340")
	theme.Colors.Chat.Self = lipgloss.Color("#BAE67E")
	theme.Colors.Chat.Connected = lipgloss.Color("#95E6CB")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#F07178")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#6E6012")
	theme.Colors.Chat.ServerMessage = lipgloss.Color("#E8CA20")

	theme.Colors.Palette.Border = lipgloss.Color("#59C2FF")
	theme.Colors.Palette.Highlight = lipgloss.Color("#59C2FF")

	theme.Colors.Nicknames = []Color{
		lipgloss.Color("#39BAE6"),
		lipgloss.Color("#FFB454"),
		lipgloss.Color("#59C2FF"),
		lipgloss.Color("#AAD94C"),
		lipgloss.Color("#95E6CB"),
		lipgloss.Color("#F07178"),
		lipgloss.Color("#FF8F40"),
		lipgloss.Color("#D2A6FF"),
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

	theme.Styles.NotifiedItem = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Sidebar.Notification).
		Bold(true)

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

func RosePineTheme() Theme {
	theme := Theme{}

	theme.Colors.Base.Background = lipgloss.Color("#191724")
	theme.Colors.Base.Foreground = lipgloss.Color("#E0DEF4")
	theme.Colors.Base.Surface = lipgloss.Color("#26233A")
	theme.Colors.Base.Border = lipgloss.Color("#403D52")
	theme.Colors.Base.Accent = lipgloss.Color("#C4A7E7")
	theme.Colors.Base.Error = lipgloss.Color("#EB6F92")
	theme.Colors.Base.Dimmed = lipgloss.Color("#6E6A86")
	theme.Colors.Base.Subtle = lipgloss.Color("#908CAA")

	theme.Colors.Sidebar.Server = lipgloss.Color("#C4A7E7")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#F6C177")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#F6C177")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#EB6F92")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#26233A")

	theme.Colors.Chat.Nickname = lipgloss.Color("#EBBCBA")
	theme.Colors.Chat.Separator = lipgloss.Color("#403D52")
	theme.Colors.Chat.Self = lipgloss.Color("#EBBCBA")
	theme.Colors.Chat.Connected = lipgloss.Color("#9CCFD8")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#EB6F92")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#6E6A86")
	theme.Colors.Chat.ServerMessage = lipgloss.Color("#F6C177")

	theme.Colors.Palette.Border = lipgloss.Color("#C4A7E7")
	theme.Colors.Palette.Highlight = lipgloss.Color("#C4A7E7")

	theme.Colors.Nicknames = []Color{
		lipgloss.Color("#EBBCBA"),
		lipgloss.Color("#F6C177"),
		lipgloss.Color("#9CCFD8"),
		lipgloss.Color("#C4A7E7"),
		lipgloss.Color("#EB6F92"),
		lipgloss.Color("#5A9FB0"),
		lipgloss.Color("#E0DEF4"),
		lipgloss.Color("#908CAA"),
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

	theme.Styles.NotifiedItem = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Sidebar.Notification).
		Bold(true)

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
