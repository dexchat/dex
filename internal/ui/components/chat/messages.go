package chat

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/irc"
)

type Message struct {
	Username  string
	Text      string
	Timestamp time.Time
	Type      irc.MessageType
}

func (m *Model) renderMessage(msg Message, width int) string {
	baseStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Background)

	timeStyle := baseStyle.
		Foreground(m.theme.Colors.Timestamp)

	styledTime := timeStyle.Render(msg.Timestamp.Format("15:04"))

	switch {
	case msg.Username == "-->" || msg.Username == "<--":
		eventStyle := baseStyle.Foreground(m.theme.Colors.ServerEventMsg)
		return baseStyle.Width(width).Render(styledTime + eventStyle.Render(" "+msg.Username+" "+msg.Text))

	case msg.Type == irc.MessageTypeConnected:
		textStyle := baseStyle.Foreground(m.theme.Colors.ConnectedMsg)
		return baseStyle.Width(width).Render(styledTime + textStyle.Render(" "+msg.Username+" "+msg.Text))

	case msg.Type == irc.MessageTypeDisconnected:
		textStyle := baseStyle.Foreground(m.theme.Colors.DisconnectedMsg)
		return baseStyle.Width(width).Render(styledTime + textStyle.Render(" "+msg.Username+" "+msg.Text))

	case msg.Type == irc.MessageTypeServer:
		serverStyle := baseStyle.Foreground(m.theme.Colors.ServerMsg)
		return baseStyle.Width(width).Render(styledTime + serverStyle.Render(" "+msg.Username+" "+msg.Text))

	default:
		nickStyle := baseStyle.
			Foreground(m.usernameColors.GetColor(msg.Username))
		styledNick := nickStyle.Render(" " + msg.Username + " ")
		styledText := renderIRCFormattedMessage(msg.Text, baseStyle)
		return baseStyle.Width(width).Render(styledTime + styledNick + styledText)
	}
}
