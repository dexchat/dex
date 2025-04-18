package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TODO: centralize all colors
// TODO: one file for each component style

var (
	PrimaryColor   = lipgloss.Color("36")
	SecondaryColor = lipgloss.Color("62")
	TextColor      = lipgloss.Color("252")
	MutedColor     = lipgloss.Color("240")
	ErrorColor     = lipgloss.Color("160")
	SuccessColor   = lipgloss.Color("76")
	WarningColor   = lipgloss.Color("214")
	HighlightColor = lipgloss.Color("226")
)

var (
	ServerInfoStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, false, false, false).
			BorderForeground(SecondaryColor).
			Padding(0, 1).
			Foreground(TextColor)

	ChannelListStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SecondaryColor).
				Padding(1).
				Foreground(TextColor)

	ChatViewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1).
			Foreground(TextColor)

	UserListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1).
			Foreground(TextColor)

	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1).
			Foreground(TextColor)

	StatusConnected = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render("connected")

	StatusDisconnected = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Render("disconnected")

	StatusConnecting = lipgloss.NewStyle().
				Foreground(WarningColor).
				Render("connecting...")

	SystemMsgStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	UserMsgStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	HighlightMsgStyle = lipgloss.NewStyle().
				Foreground(HighlightColor)
)

func UsernameStyle(username string) string {
	colorMap := map[string]lipgloss.Color{
		"you":     PrimaryColor,
		"admin":   ErrorColor,
		"op":      SuccessColor,
		"voice":   WarningColor,
		"default": TextColor,
	}

	color := colorMap["default"]
	for role, roleColor := range colorMap {
		if role != "default" && (username == role || strings.HasPrefix(username, role)) {
			color = roleColor
			break
		}
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(username)
}
