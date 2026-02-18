package ui

import (
	"strings"
	"time"

	"github.com/vaaleyard/dex/internal/history"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
	"github.com/vaaleyard/dex/internal/ui/components/users"
)

// BufferKey is the identifier for any buffer (servers or channels) in the app
type BufferKey string

func makeBufferKey(server, channel string) BufferKey {
	// IRC channel names are case-insensitive
	return BufferKey(strings.ToLower(server) + ":" + strings.ToLower(channel))
}

type Buffer struct {
	Key    BufferKey
	Server string
	Buffer string // buffer can be a channel or a PM

	Chat    chat.Model
	Users   users.Model
	History *history.Log
}

func (b *Buffer) isValid() bool {
	return b.Server != "" && b.Buffer != ""
}

func (b *Buffer) LoadHistory() {
	entries := b.History.Entries()
	msgs := make([]chat.Message, len(entries))
	for i, entry := range entries {
		msgs[i] = chat.Message{
			Timestamp: time.Unix(0, entry.ServerTime),
			Username:  entry.Username,
			Text:      entry.Text,
			Type:      irc.MessageType(entry.Type),
		}
	}
	b.Chat.AddMessages(msgs)
}
