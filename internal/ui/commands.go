package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/commands"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/components/chat"
	"github.com/lrstanley/girc"
)

func (m *Model) handleCommand(buffer *Buffer, command commands.Command) tea.Cmd {
	switch command.Name {
	case "help":
		m.showHelp()
	case "join":
		m.handleJoinCommand(buffer, command.Args)
	case "leave":
		m.handleLeaveCommand(buffer, command.Args)
	case "part":
		m.handleLeaveCommand(buffer, command.Args)
	case "close":
		return m.handleCloseCommand(buffer, command.Args)
	case "list":
		m.handleListCommand(buffer, command.Args)
	case "msg":
		m.handleMsgCommand(buffer, command.Args)
	default:
		m.addCommandError(buffer, "unknown command: /"+command.Name)
	}
	return nil
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

func (m *Model) handleCloseCommand(buffer *Buffer, args []string) tea.Cmd {
	if len(args) != 0 {
		m.addCommandError(buffer, "usage: /close")
		return nil
	}
	if buffer.Buffer == "" || girc.IsValidChannel(buffer.Buffer) {
		m.addCommandError(buffer, "error: /close is only available in a private message")
		return nil
	}

	return m.removeBuffer(buffer, true)
}

func (m *Model) handleListCommand(buffer *Buffer, args []string) {
	if len(args) > 1 {
		m.addCommandError(buffer, "usage: /list [channel]")
		return
	}

	channel := ""
	if len(args) == 1 {
		channel = args[0]
	}
	if m.ircClientManager != nil {
		go m.ircClientManager.List(buffer.Server, channel)
	}
}

func (m *Model) handleMsgCommand(buffer *Buffer, args []string) {
	if len(args) < 2 {
		m.addCommandError(buffer, "usage: /msg <user> <message>")
		return
	}

	target := args[0]
	message := strings.Join(args[1:], " ")
	privateBuffer, createCmd := m.getOrCreateBuffer(buffer.Server, target)
	if privateBuffer == nil {
		return
	}
	if createCmd != nil {
		m.channels, _ = m.channels.Update(createCmd())
	}

	// A /msg starts a private conversation immediately. Unlike /join, no
	// server event is needed to discover this buffer.
	m.activeBuffer = privateBuffer.Key
	m.clearBufferActivity(privateBuffer.Key)
	privateBuffer.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
	if m.directMessages.Add(buffer.Server, target) {
		m.directMessagesDirty = true
	}

	now := time.Now()
	privateBuffer.Chat.AddMessage(chat.Message{
		Timestamp: now,
		Username:  privateBuffer.Chat.Nickname(),
		Text:      message,
	})
	if m.ircClientManager != nil {
		go m.ircClientManager.Send(buffer.Server, target, message)
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
