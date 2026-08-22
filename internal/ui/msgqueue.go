package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/history"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui/components/channels"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
)

type flushChatMsg struct{}

// Keep replay work below a frame's worth of UI processing. ZNC can deliver a
// very large playback burst in one IRC batch; processing it all in one Update
// starves rendering and input until the burst is exhausted.
const maxPlaybackMessagesPerUpdate = 50

const maxNotificationAge = 30 * time.Second

const notificationNoticeDuration = 4 * time.Second

type notificationNoticeExpiredMsg struct {
	version uint64
}

func playbackChunk(messages irc.BufferNewMessageBatchMsg) (current, remaining irc.BufferNewMessageBatchMsg) {
	if len(messages) <= maxPlaybackMessagesPerUpdate {
		return messages, nil
	}
	return messages[:maxPlaybackMessagesPerUpdate], messages[maxPlaybackMessagesPerUpdate:]
}

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
	m.getActiveBuffer().Chat.FlushQueue()
}

// processIncomingMessage handles a single incoming IRC message:
// deduplicates, inserts into history, and queues for rendering
func (m *Model) processIncomingMessage(msg irc.BufferNewMessageMsg) tea.Cmd {
	buf, createCmd := m.getOrCreateBuffer(msg.Server, msg.Buffer)
	if buf == nil {
		return createCmd
	}
	if msg.DirectMessage && m.directMessages.Add(msg.Server, msg.Buffer) {
		m.directMessagesDirty = true
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
	if msg.Type == irc.MessageTypeNormal {
		buf.recordLatestMessage(logEntry)
	}

	if buf.History.IsDuplicate(logEntry) {
		return createCmd
	}

	buf.History.Insert(logEntry)

	var notificationCmd tea.Cmd

	if !msg.OwnEcho {
		buf.Chat.QueueMessage(chat.Message{
			Timestamp: msg.Timestamp,
			Username:  msg.From,
			Text:      msg.Text,
			Type:      msg.Type,
		})
		m.updateActivityForMessage(buf, msg)
		notificationNoticeCmd := m.showNotificationNotice(buf, msg)
		attentionPulseCmd := m.startAttentionPulse(buf, msg)

		if m.shouldSoundNotification(buf, msg) {
			notificationCmd = m.soundNotificationCmd()
		}
		return tea.Batch(createCmd, notificationCmd, notificationNoticeCmd, attentionPulseCmd)
	}

	return tea.Batch(createCmd, notificationCmd)
}

func (m *Model) updateActivityForMessage(buf *Buffer, msg irc.BufferNewMessageMsg) {
	if msg.Type != irc.MessageTypeNormal {
		return
	}
	if buf.Key == m.activeBuffer {
		return
	}
	if m.messageIsRead(buf, msg.Timestamp) {
		return
	}

	buf.UnreadCount++
	if m.config.NotifiesChannel(buf.Server, msg.Buffer) {
		buf.NotificationCount++
	}
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
	m.markBufferRead(buf)
	if buf.UnreadCount == 0 && buf.MentionCount == 0 {
		return
	}

	buf.UnreadCount = 0
	buf.NotificationCount = 0
	buf.MentionCount = 0
	m.updateChannelActivity(buf)
}

func (m *Model) messageIsRead(buf *Buffer, timestamp time.Time) bool {
	if m.readState == nil {
		return false
	}
	marker, ok := m.readState.Marker(buf.Server, buf.Buffer)
	return ok && timestamp.UnixNano() <= marker.ServerTime
}

func (m *Model) markBufferRead(buf *Buffer) {
	if buf.latestMessage.ServerTime == 0 {
		return
	}
	if m.readState == nil {
		m.readState = &history.ReadState{}
	}

	marker, exists := m.readState.Marker(buf.Server, buf.Buffer)
	if exists && marker.ServerTime >= buf.latestMessage.ServerTime {
		return
	}

	m.readState.MarkRead(buf.Server, buf.Buffer, buf.latestMessage)
	m.readStateDirty = true
}

func (b *Buffer) recordLatestMessage(entry history.LogEntry) {
	if entry.ServerTime < b.latestMessage.ServerTime {
		return
	}
	if entry.ServerTime == b.latestMessage.ServerTime && b.latestMessage.MsgID != "" {
		return
	}

	b.latestMessage = history.ReadMarker{ServerTime: entry.ServerTime}
	if entry.MsgID != nil {
		b.latestMessage.MsgID = *entry.MsgID
	}
}

func (m *Model) updateChannelActivity(buf *Buffer) {
	settings := m.badgeSettingsForServer(buf.Server)
	unreadCount := buf.UnreadCount
	notificationCount := buf.NotificationCount
	mentionCount := buf.MentionCount
	if !settings.Unread {
		unreadCount = 0
		notificationCount = 0
	}
	if !settings.Mention {
		mentionCount = 0
	}

	m.channels, _ = m.channels.Update(channels.ActivityUpdateMsg{
		Server:            buf.Server,
		Buffer:            buf.Buffer,
		UnreadCount:       unreadCount,
		NotificationCount: notificationCount,
		MentionCount:      mentionCount,
	})
}

func (m *Model) showNotificationNotice(buf *Buffer, msg irc.BufferNewMessageMsg) tea.Cmd {
	if msg.Type != irc.MessageTypeNormal || buf.Key == m.activeBuffer ||
		!m.config.NotifiesChannel(buf.Server, msg.Buffer) {
		return nil
	}
	if !msg.Timestamp.IsZero() && m.now().Sub(msg.Timestamp) > maxNotificationAge {
		return nil
	}

	m.notificationNoticeVersion++
	version := m.notificationNoticeVersion
	m.channels = m.channels.ShowNotificationNotice(buf.Server, msg.Buffer, msg.From)
	return tea.Tick(notificationNoticeDuration, func(_ time.Time) tea.Msg {
		return notificationNoticeExpiredMsg{version: version}
	})
}

func (m *Model) startAttentionPulse(buf *Buffer, msg irc.BufferNewMessageMsg) tea.Cmd {
	if msg.Type != irc.MessageTypeNormal || buf.Key == m.activeBuffer {
		return nil
	}
	if !msg.Timestamp.IsZero() && m.now().Sub(msg.Timestamp) > maxNotificationAge {
		return nil
	}
	if buf.MentionCount == 0 &&
		!m.config.NotifiesChannel(buf.Server, msg.Buffer) &&
		!messageMentionsNick(msg.Text, buf.Chat.Nickname()) {
		return nil
	}

	var cmd tea.Cmd
	m.channels, cmd = m.channels.StartNotificationPulse(buf.Server, msg.Buffer)
	return cmd
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
	return chat.MessageMentionsNick(message, nick)
}

func (m *Model) shouldSoundNotification(buf *Buffer, msg irc.BufferNewMessageMsg) bool {
	if !m.config.Notifications.Sound {
		return false
	}

	if msg.Type != irc.MessageTypeNormal {
		return false
	}

	if !msg.Timestamp.IsZero() &&
		m.now().Sub(msg.Timestamp) > maxNotificationAge {
		return false
	}

	if msg.OwnEcho {
		return false
	}

	// No need for notification in the current open buffer
	if buf.Key == m.activeBuffer {
		return false
	}

	if msg.DirectMessage {
		if m.config.IgnoresDirectMessageFrom(buf.Server, msg.From) {
			return false
		}
		return m.config.Notifications.Events[config.NotificationDirectMessage]
	}

	if m.config.NotifiesChannel(buf.Server, msg.Buffer) {
		return true
	}

	if m.config.Notifications.Events[config.NotificationMention] &&
		messageMentionsNick(msg.Text, buf.Chat.Nickname()) {
		return true
	}

	return false
}

func (m *Model) soundNotificationCmd() tea.Cmd {
	now := m.now()

	if !m.lastSoundAt.IsZero() &&
		now.Sub(m.lastSoundAt) < m.config.Notifications.Cooldown {
		return nil
	}

	m.lastSoundAt = now
	return tea.Raw("\a")
}
