package ui

import (
	"strings"
	"time"

	"github.com/vaaleyard/dex/internal/commands"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
)

func (m *Model) handleCommand(buffer *Buffer, command commands.Command) {
	switch command.Name {
	case "join":
		m.handleJoinCommand(buffer, command.Args)
	case "leave":
		m.handleLeaveCommand(buffer, command.Args)
	default:
		m.addCommandError(buffer, "unknown command: /"+command.Name)
	}
}

func (m *Model) handleJoinCommand(buffer *Buffer, args []string) {
	if len(args) < 1 || len(args) > 2 {
		m.addCommandError(buffer, "usage: /join <channel> [key]")
		return
	}

	key := ""
	if len(args) == 2 {
		key = args[1]
	}
	if m.ircClientManager != nil {
		go m.ircClientManager.Join(buffer.Server, args[0], key)
	}
}

func (m *Model) handleLeaveCommand(buffer *Buffer, args []string) {
	if m.ircClientManager != nil {
		go m.ircClientManager.Part(buffer.Server, buffer.Buffer, strings.Join(args, " "))
	}
}

func (m *Model) addCommandError(buffer *Buffer, text string) {
	buffer.Chat.AddMessage(chat.Message{
		Timestamp: time.Now(),
		Username:  "--",
		Text:      text,
		Type:      irc.MessageTypeServer,
	})
}
