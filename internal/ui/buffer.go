package ui

import (
	"time"

	"github.com/dexchat/dex/internal/history"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/components/chat"
	"github.com/dexchat/dex/internal/ui/components/users"
)

// BufferKey is the identifier for any buffer (servers or channels) in the app
type BufferKey string

func makeBufferKey(server, channel string) BufferKey {
	// IRC channel names are case-insensitive
	return BufferKey(history.NameKey(server) + ":" + history.NameKey(channel))
}

// historyState tracks whether a buffer's stored history is in memory.
type historyState int

const (
	// historyNotLoaded means saves must merge with the stored file first.
	historyNotLoaded historyState = iota
	historyLoading
	historyLoaded
	// historyUnreadable disables saves so the unreadable stored file is never
	// replaced. Selecting the buffer again retries the load.
	historyUnreadable
)

type Buffer struct {
	Key    BufferKey
	Server string
	Buffer string // buffer can be a channel or a PM

	Chat    chat.Model
	Users   users.Model
	History *history.Log

	// members is the latest IRC snapshot. Inactive buffers retain data without
	// paying the cost of rendering every user list during ZNC playback.
	members        []string
	memberPrefixes string

	// Channel history is intentionally loaded on demand. Starting one disk job
	// per channel during a ZNC replay can monopolize the machine at startup.
	historyState  historyState
	latestMessage history.ReadMarker

	UnreadCount       int
	NotificationCount int
	MentionCount      int
}

func (b *Buffer) isValid() bool {
	return b.Server != "" && b.Buffer != ""
}

func (b *Buffer) LoadHistory() {
	messages := make([]chat.Message, 0, len(b.History.Entries()))
	for _, entry := range b.History.Entries() {
		messages = append(messages, chat.Message{
			Timestamp: time.Unix(0, entry.ServerTime),
			Username:  entry.Username,
			Text:      entry.Text,
			Action:    entry.Action,
			Type:      irc.MessageType(entry.Type),
		})
	}
	b.Chat.ReplaceMessages(messages)
}
