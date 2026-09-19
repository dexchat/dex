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
	theme.Colors.Base.Dimmed = lipgloss.Color("#89919B")
	theme.Colors.Base.Subtle = lipgloss.Color("#75808A")

	theme.Colors.Sidebar.Server = lipgloss.Color("#59C2FF")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#73D0FF")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#FFB454")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#F07178")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#253340")

	theme.Colors.Chat.Nickname = lipgloss.Color("#59C2FF")
	theme.Colors.Chat.InactiveNickname = lipgloss.Color("#7D7D7D")
	theme.Colors.Chat.Mention = lipgloss.Color("#F07178")
	theme.Colors.Chat.Separator = lipgloss.Color("#253340")
	theme.Colors.Chat.Self = lipgloss.Color("#BAE67E")
	theme.Colors.Chat.Connected = lipgloss.Color("#95E6CB")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#F07178")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#75808A")
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

func DraculaTheme() Theme {
	theme := Theme{}

	theme.Colors.Base.Background = lipgloss.Color("#282A36")
	theme.Colors.Base.Foreground = lipgloss.Color("#F8F8F2")
	theme.Colors.Base.Surface = lipgloss.Color("#343746")
	theme.Colors.Base.Border = lipgloss.Color("#6272A4")
	theme.Colors.Base.Accent = lipgloss.Color("#BD93F9")
	theme.Colors.Base.Error = lipgloss.Color("#FF5555")
	theme.Colors.Base.Dimmed = lipgloss.Color("#B8B8C2")
	theme.Colors.Base.Subtle = lipgloss.Color("#B8B8C2")

	theme.Colors.Sidebar.Server = lipgloss.Color("#8BE9FD")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#D6ACFF")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#F1FA8C")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#FF92DF")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#44475A")

	theme.Colors.Chat.Nickname = lipgloss.Color("#8BE9FD")
	theme.Colors.Chat.InactiveNickname = lipgloss.Color("#B8B8C2")
	theme.Colors.Chat.Mention = lipgloss.Color("#FF79C6")
	theme.Colors.Chat.Separator = lipgloss.Color("#6272A4")
	theme.Colors.Chat.Self = lipgloss.Color("#50FA7B")
	theme.Colors.Chat.Connected = lipgloss.Color("#50FA7B")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#FF5555")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#B8B8C2")
	theme.Colors.Chat.ServerMessage = lipgloss.Color("#FFB86C")

	theme.Colors.Palette.Border = lipgloss.Color("#BD93F9")
	theme.Colors.Palette.Highlight = lipgloss.Color("#BD93F9")

	theme.Colors.Nicknames = []Color{
		lipgloss.Color("#FF5555"),
		lipgloss.Color("#FFB86C"),
		lipgloss.Color("#F1FA8C"),
		lipgloss.Color("#50FA7B"),
		lipgloss.Color("#8BE9FD"),
		lipgloss.Color("#BD93F9"),
		lipgloss.Color("#FF79C6"),
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

func GruvboxDarkTheme() Theme {
	theme := Theme{}

	theme.Colors.Base.Background = lipgloss.Color("#1D2021")
	theme.Colors.Base.Foreground = lipgloss.Color("#EBDBB2")
	theme.Colors.Base.Surface = lipgloss.Color("#282828")
	theme.Colors.Base.Border = lipgloss.Color("#504945")
	theme.Colors.Base.Accent = lipgloss.Color("#8EC07C")
	theme.Colors.Base.Error = lipgloss.Color("#FB4934")
	theme.Colors.Base.Dimmed = lipgloss.Color("#A89984")
	theme.Colors.Base.Subtle = lipgloss.Color("#A89984")

	theme.Colors.Sidebar.Server = lipgloss.Color("#8EC07C")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#83A598")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#FABD2F")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#FE8019")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#282828")

	theme.Colors.Chat.Nickname = lipgloss.Color("#8EC07C")
	theme.Colors.Chat.InactiveNickname = lipgloss.Color("#A89984")
	theme.Colors.Chat.Mention = lipgloss.Color("#FB4934")
	theme.Colors.Chat.Separator = lipgloss.Color("#504945")
	theme.Colors.Chat.Self = lipgloss.Color("#B8BB26")
	theme.Colors.Chat.Connected = lipgloss.Color("#8EC07C")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#FB4934")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#A89984")
	theme.Colors.Chat.ServerMessage = lipgloss.Color("#FABD2F")

	theme.Colors.Palette.Border = lipgloss.Color("#8EC07C")
	theme.Colors.Palette.Highlight = lipgloss.Color("#8EC07C")

	theme.Colors.Nicknames = []Color{
		lipgloss.Color("#FB4934"),
		lipgloss.Color("#B8BB26"),
		lipgloss.Color("#FABD2F"),
		lipgloss.Color("#83A598"),
		lipgloss.Color("#D3869B"),
		lipgloss.Color("#8EC07C"),
		lipgloss.Color("#FE8019"),
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
	theme.Colors.Base.Dimmed = lipgloss.Color("#908CAA")
	theme.Colors.Base.Subtle = lipgloss.Color("#908CAA")

	theme.Colors.Sidebar.Server = lipgloss.Color("#C4A7E7")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#F6C177")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#F6C177")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#EB6F92")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#26233A")

	theme.Colors.Chat.Nickname = lipgloss.Color("#EBBCBA")
	theme.Colors.Chat.InactiveNickname = lipgloss.Color("#858585")
	theme.Colors.Chat.Mention = lipgloss.Color("#EB6F92")
	theme.Colors.Chat.Separator = lipgloss.Color("#403D52")
	theme.Colors.Chat.Self = lipgloss.Color("#EBBCBA")
	theme.Colors.Chat.Connected = lipgloss.Color("#9CCFD8")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#EB6F92")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#908CAA")
	theme.Colors.Chat.ServerMessage = lipgloss.Color("#F6C177")

	theme.Colors.Palette.Border = lipgloss.Color("#C4A7E7")
	theme.Colors.Palette.Highlight = lipgloss.Color("#C4A7E7")

	theme.Colors.Nicknames = []Color{
		lipgloss.Color("#EBBCBA"),
		lipgloss.Color("#F6C177"),
		lipgloss.Color("#9CCFD8"),
		lipgloss.Color("#C4A7E7"),
		lipgloss.Color("#EB6F92"),
		lipgloss.Color("#E0DEF4"),
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

func SolarizedLightTheme() Theme {
	theme := Theme{}

	theme.Colors.Base.Background = lipgloss.Color("#FDF6E3")
	theme.Colors.Base.Foreground = lipgloss.Color("#4E646B")
	theme.Colors.Base.Surface = lipgloss.Color("#EEE8D5")
	theme.Colors.Base.Border = lipgloss.Color("#93A1A1")
	theme.Colors.Base.Accent = lipgloss.Color("#1B6EA3")
	theme.Colors.Base.Error = lipgloss.Color("#C23130")
	theme.Colors.Base.Dimmed = lipgloss.Color("#556B72")
	theme.Colors.Base.Subtle = lipgloss.Color("#556B72")

	theme.Colors.Sidebar.Server = lipgloss.Color("#1A7473")
	theme.Colors.Sidebar.Unread = lipgloss.Color("#1B6EA3")
	theme.Colors.Sidebar.Notification = lipgloss.Color("#766813")
	theme.Colors.Sidebar.Mention = lipgloss.Color("#B1471A")
	theme.Colors.Sidebar.Selection = lipgloss.Color("#EEE8D5")

	theme.Colors.Chat.Nickname = lipgloss.Color("#1B6EA3")
	theme.Colors.Chat.InactiveNickname = lipgloss.Color("#556B72")
	theme.Colors.Chat.Mention = lipgloss.Color("#B53477")
	theme.Colors.Chat.Separator = lipgloss.Color("#93A1A1")
	theme.Colors.Chat.Self = lipgloss.Color("#557113")
	theme.Colors.Chat.Connected = lipgloss.Color("#1A7473")
	theme.Colors.Chat.Disconnected = lipgloss.Color("#C23130")
	theme.Colors.Chat.UserEvents = lipgloss.Color("#556B72")
	theme.Colors.Chat.ServerMessage = lipgloss.Color("#B1471A")

	theme.Colors.Palette.Border = lipgloss.Color("#1B6EA3")
	theme.Colors.Palette.Highlight = lipgloss.Color("#1B6EA3")

	theme.Colors.Nicknames = []Color{
		lipgloss.Color("#766813"),
		lipgloss.Color("#B1471A"),
		lipgloss.Color("#C23130"),
		lipgloss.Color("#B53477"),
		lipgloss.Color("#5764A9"),
		lipgloss.Color("#1B6EA3"),
		lipgloss.Color("#1A7473"),
		lipgloss.Color("#557113"),
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
