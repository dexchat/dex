package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/history"
	"github.com/vaaleyard/dex/internal/irc"
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
	}

	return createCmd
}
