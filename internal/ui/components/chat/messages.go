package chat

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	Username string
	Text     string
	Time     string
}

// functions to display raw messages for the POC
func parseMessage(rawMsg string) Message {
	parts := strings.Fields(rawMsg)
	if len(parts) < 2 {
		return Message{}
	}
	return Message{
		Time:     "",
		Username: parts[0],
		Text:     strings.Join(parts[1:], " "),
	}
}

func (m *Model) renderMessage(msg Message, width int) string {
	baseStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Background)

	nickStyle := baseStyle.
		Foreground(m.usernameColors.GetColor(msg.Username))

	styledNick := nickStyle.Render(msg.Username)
	styledText := baseStyle.Render(" " + msg.Text)

	return baseStyle.Width(width).Render(styledNick + styledText)
}
