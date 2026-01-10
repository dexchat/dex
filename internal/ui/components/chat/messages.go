package chat

import (
	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	Username string
	Text     string
	Time     string
}

func (m *Model) renderMessage(msg Message, width int) string {
	baseStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Background)

	timeStyle := baseStyle.
		Foreground(m.theme.Colors.Timestamp)

	nickStyle := baseStyle.
		Foreground(m.usernameColors.GetColor(msg.Username))

	styledTime := timeStyle.Render(msg.Time)
	styledNick := nickStyle.Render(" " + msg.Username)
	styledText := baseStyle.Render(" " + msg.Text)

	return baseStyle.Width(width).Render(styledTime + styledNick + styledText)
}
