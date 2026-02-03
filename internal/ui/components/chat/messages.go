package chat

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	Username  string
	Text      string
	Timestamp time.Time
}

func (m *Model) renderMessage(msg Message, width int) string {
	baseStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Background)

	timeStyle := baseStyle.
		Foreground(m.theme.Colors.Timestamp)

	nickStyle := baseStyle.
		Foreground(m.usernameColors.GetColor(msg.Username))

	styledTime := timeStyle.Render(msg.Timestamp.Format("15:04"))
	styledNick := nickStyle.Render(" " + msg.Username + " ")
	styledText := renderIRCFormattedMessage(msg.Text, baseStyle)

	return baseStyle.Width(width).Render(styledTime + styledNick + styledText)
}
