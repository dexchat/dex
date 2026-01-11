package ui

import (
	"strings"

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
	Key     BufferKey
	Server  string
	Channel string

	Chat  chat.Model
	Users users.Model
}
