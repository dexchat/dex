package styles

import "github.com/charmbracelet/lipgloss"

// TODO: Improve the name of each color

// Ayu
func AyuDarkTheme() Theme {
	theme := Theme{}

	theme.Colors.Background = lipgloss.Color("#0D1017")
	theme.Colors.LighterBackground = lipgloss.Color("#131721")
	theme.Colors.Text = lipgloss.Color("#B3B1AD")
	theme.Colors.Accent = lipgloss.Color("#59C2FF")
	theme.Colors.SidebarBg = lipgloss.Color("#131721")
	theme.Colors.InputBg = lipgloss.Color("#1F2430")
	theme.Colors.BorderColor = lipgloss.Color("#253340")
	theme.Colors.SelfMsg = lipgloss.Color("#BAE67E")
	theme.Colors.SystemMsg = lipgloss.Color("#FFB454")
	theme.Colors.ErrorMsg = lipgloss.Color("#FF3333")
	theme.Colors.Timestamp = lipgloss.Color("#626A73")
	theme.Colors.MentionColor = lipgloss.Color("#F07178")
	theme.Colors.UnreadColor = lipgloss.Color("#73D0FF")
	theme.Colors.StatusBg = lipgloss.Color("#253340")
	theme.Colors.StatusText = lipgloss.Color("#B3B1AD")
	theme.Colors.Usernames = []lipgloss.Color{
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
		Background(theme.Colors.Background).
		Foreground(theme.Colors.Text)

	theme.Styles.Sidebar = lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Foreground(theme.Colors.Text).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Colors.Background).
		BorderBackground(theme.Colors.Background)

	theme.Styles.SelectedItem = lipgloss.NewStyle().
		Background(lipgloss.Color("#E6B450")).
		Foreground(lipgloss.Color("#0A0E14"))

	theme.Styles.UnreadItem = lipgloss.NewStyle().
		Foreground(theme.Colors.UnreadColor)

	theme.Styles.ChatArea = lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Foreground(theme.Colors.Text).
		BorderForeground(theme.Colors.Background).
		BorderBackground(theme.Colors.Background).
		Border(lipgloss.NormalBorder())

	theme.Styles.OwnMessage = lipgloss.NewStyle().
		Foreground(theme.Colors.SelfMsg)

	theme.Styles.SystemMessage = lipgloss.NewStyle().
		Foreground(theme.Colors.SystemMsg).
		Background(theme.Colors.Background)

	theme.Styles.ErrorMessage = lipgloss.NewStyle().
		Foreground(theme.Colors.ErrorMsg)

	theme.Styles.Timestamp = lipgloss.NewStyle().
		Foreground(theme.Colors.Timestamp)

	theme.Styles.Mention = lipgloss.NewStyle().
		Foreground(theme.Colors.MentionColor).
		Bold(true)

	theme.Styles.InputField = lipgloss.NewStyle().
		Background(theme.Colors.LighterBackground).
		Foreground(theme.Colors.Text)

	theme.Styles.StatusLine = lipgloss.NewStyle().
		Background(theme.Colors.StatusBg).
		Foreground(theme.Colors.StatusText)

	theme.Styles.Usernames = lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Foreground(theme.Colors.Usernames[7])

	return theme
}
