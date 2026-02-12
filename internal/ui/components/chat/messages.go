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
		Background(m.theme.Colors.Base.Background)

	timeStyle := baseStyle.
		Foreground(m.theme.Colors.Base.Dimmed)

	styledTime := timeStyle.Render(msg.Timestamp.Format("15:04"))

	styledText := renderIRCFormattedMessage(msg.Text, baseStyle)

	switch {
	case msg.Username == "-->" || msg.Username == "<--":
		eventStyle := baseStyle.Foreground(m.theme.Colors.Chat.UserEvents)
		return baseStyle.Width(width).Render(styledTime + eventStyle.Render(" "+msg.Username+" ") + styledText)

	case msg.Type == irc.MessageTypeConnected:
		textStyle := baseStyle.Foreground(m.theme.Colors.Chat.Connected)
		return baseStyle.Width(width).Render(styledTime + textStyle.Render(" "+msg.Username+" ") + styledText)

	case msg.Type == irc.MessageTypeDisconnected:
		textStyle := baseStyle.Foreground(m.theme.Colors.Chat.Disconnected)
		return baseStyle.Width(width).Render(styledTime + textStyle.Render(" "+msg.Username+" ") + styledText)

	case msg.Type == irc.MessageTypeServer:
		serverStyle := baseStyle.Foreground(m.theme.Colors.Chat.ServerMessage)
		return baseStyle.Width(width).Render(styledTime + serverStyle.Render(" "+msg.Username+" ") + styledText)

	default:
		nickStyle := baseStyle.
			Foreground(m.usernameColors.GetColor(msg.Username))
		styledNick := nickStyle.Render(" " + msg.Username + " ")
		return baseStyle.Width(width).Render(styledTime + styledNick + styledText)
	}
}
