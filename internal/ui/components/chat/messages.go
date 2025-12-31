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
	styledNick := lipgloss.NewStyle().
		Foreground(m.usernameColors.GetColor(msg.Username)).
		Render(msg.Username)

	styledMsg := lipgloss.NewStyle().
		Background(m.theme.Colors.Background).
		Width(width).
		Render(" " + msg.Text)

	return styledNick + styledMsg
}
