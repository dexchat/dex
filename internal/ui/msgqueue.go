package ui

import (
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/history"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui/components/channels"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
)

type flushChatMsg struct{}

func (m *Model) scheduleFlush() tea.Cmd {
	if m.flushPending {
		return nil
	}
	m.flushPending = true
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return flushChatMsg{}
	})
}

func (m *Model) flushAllChats() {
	m.flushPending = false
	for _, buf := range m.buffers {
		buf.Chat.FlushQueue()
	}
}

// processIncomingMessage handles a single incoming IRC message:
// deduplicates, inserts into history, and queues for rendering
func (m *Model) processIncomingMessage(msg irc.BufferNewMessageMsg) tea.Cmd {
	buf, createCmd := m.getOrCreateBuffer(msg.Server, msg.Buffer)
	if buf == nil {
		return createCmd
	}

	var msgID *string
	if msg.MsgID != "" {
		msgID = &msg.MsgID
	}

	logEntry := history.LogEntry{
		ReceivedAt: time.Now().UnixNano(),
		ServerTime: msg.Timestamp.UnixNano(),
		MsgID:      msgID,
		Username:   msg.From,
		Text:       msg.Text,
		Type:       int(msg.Type),
	}

	if buf.History.IsDuplicate(logEntry) {
		return createCmd
	}

	buf.History.Insert(logEntry)

	if !msg.OwnEcho {
		buf.Chat.QueueMessage(chat.Message{
			Timestamp: msg.Timestamp,
			Username:  msg.From,
			Text:      msg.Text,
			Type:      msg.Type,
		})
		m.updateActivityForMessage(buf, msg)
	}

	return createCmd
}

func (m *Model) updateActivityForMessage(buf *Buffer, msg irc.BufferNewMessageMsg) {
	if buf.Key == m.activeBuffer {
		return
	}

	buf.UnreadCount++
	if messageMentionsNick(msg.Text, buf.Chat.Nickname()) {
		buf.MentionCount++
	}
	m.updateChannelActivity(buf)
}

func (m *Model) clearBufferActivity(key BufferKey) {
	buf := m.buffers[key]
	if buf == nil {
		return
	}
	if buf.UnreadCount == 0 && buf.MentionCount == 0 {
		return
	}

	buf.UnreadCount = 0
	buf.MentionCount = 0
	m.updateChannelActivity(buf)
}

func (m *Model) updateChannelActivity(buf *Buffer) {
	settings := m.badgeSettingsForServer(buf.Server)
	unreadCount := buf.UnreadCount
	mentionCount := buf.MentionCount
	if !settings.Unread {
		unreadCount = 0
	}
	if !settings.Mention {
		mentionCount = 0
	}

	m.channels, _ = m.channels.Update(channels.ActivityUpdateMsg{
		Server:       buf.Server,
		Buffer:       buf.Buffer,
		UnreadCount:  unreadCount,
		MentionCount: mentionCount,
	})
}

func (m *Model) badgeSettingsForServer(serverName string) config.BadgeSettings {
	for _, server := range m.config.Servers {
		if strings.EqualFold(server.Name, serverName) {
			return m.config.ServerBadgeSettings(server)
		}
	}
	return m.config.ServerBadgeSettings(nil)
}

// messageMentionsNick matches the a given nick in a text message, so "nick" counts in
// "hey nick:" but not inside larger nick-like strings such as "supernick"
func messageMentionsNick(message, nick string) bool {
	if nick == "" {
		return false
	}

	message = strings.ToLower(message)
	nick = strings.ToLower(nick)

	for start := 0; start < len(message); {
		idx := strings.Index(message[start:], nick)
		if idx == -1 {
			return false
		}

		idx += start
		before := idx - 1
		after := idx + len(nick)
		if (before < 0 || !isNickChar(rune(message[before]))) && (after >= len(message) || !isNickChar(rune(message[after]))) {
			return true
		}
		start = idx + len(nick)
	}

	return false
}

func isNickChar(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	return strings.ContainsRune("_-[]\\`^{}|", r)
}
