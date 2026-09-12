package chat

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/irc"
)

type ChannelMembers interface {
	HasUser(nick string) bool
	GetUserPrefix(nick string) string
	Nicknames() []string
}

type Message struct {
	Username  string
	Text      string
	Timestamp time.Time
	Type      irc.MessageType
}

func (m *Model) renderMessage(msg Message, width int) string {
	baseStyle := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Base.Foreground).
		Background(m.theme.Colors.Base.Background)

	timeStyle := baseStyle.
		Foreground(m.theme.Colors.Base.Dimmed)
	styledTime := timeStyle.Render(msg.Timestamp.Format("15:04"))

	switch {
	case msg.Username == "-->" || msg.Username == "<--":
		eventStyle := baseStyle.Foreground(m.theme.Colors.Chat.UserEvents)
		return baseStyle.Width(width).Render(styledTime + eventStyle.Render(" "+msg.Username+" ") + eventStyle.Render(msg.Text))

	case msg.Type == irc.MessageTypeConnected:
		textStyle := baseStyle.Foreground(m.theme.Colors.Chat.Connected)
		return baseStyle.Width(width).Render(styledTime + textStyle.Render(" "+msg.Username+" ") + textStyle.Render(msg.Text))

	case msg.Type == irc.MessageTypeDisconnected:
		textStyle := baseStyle.Foreground(m.theme.Colors.Chat.Disconnected)
		return baseStyle.Width(width).Render(styledTime + textStyle.Render(" "+msg.Username+" ") + textStyle.Render(msg.Text))

	case msg.Type == irc.MessageTypeServer:
		serverStyle := baseStyle.Foreground(m.theme.Colors.Chat.ServerMessage)
		return baseStyle.Width(width).Render(styledTime + serverStyle.Render(" "+msg.Username+" ") + renderIRCFormattedMessage(msg.Text, serverStyle))

	default:
		mentionStyle := baseStyle.Foreground(m.theme.Colors.Chat.Mention).Bold(true)
		styledText := renderIRCFormattedMessage(msg.Text, baseStyle)
		if !strings.EqualFold(msg.Username, m.nickname) {
			styledText, _ = renderIRCFormattedMessageWithMentions(msg.Text, m.nickname, baseStyle, mentionStyle)
		}
		nickColor := m.usernameColors.GetColor(msg.Username)
		prefix := ""
		if m.channelMembers != nil {
			if !m.channelMembers.HasUser(msg.Username) {
				nickColor = m.theme.Colors.Base.Dimmed
			} else {
				prefix = m.channelMembers.GetUserPrefix(msg.Username)
			}
		}
		nickStyle := baseStyle.Foreground(nickColor)
		styledNick := nickStyle.Render(" " + prefix + msg.Username + " ")
		return baseStyle.Width(width).Render(styledTime + styledNick + styledText)
	}
}
