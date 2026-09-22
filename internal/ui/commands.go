package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/commands"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/components/chat"
)

func (m *Model) handleCommand(buffer *Buffer, command commands.Command) tea.Cmd {
	switch command.Name {
	case "help":
		m.showHelp()
	case "join":
		return m.handleJoinCommand(buffer, command.Args)
	case "leave":
		return m.handleLeaveCommand(buffer, command.Args)
	case "part":
		return m.handleLeaveCommand(buffer, command.Args)
	case "close":
		return m.handleCloseCommand(buffer, command.Args)
	case "list":
		return m.handleListCommand(buffer, command.Args)
	case "msg":
		return m.handleMsgCommand(buffer, command.Args)
	default:
		m.addCommandError(buffer, "unknown command: /"+command.Name)
	}
	return nil
}

func (m *Model) handleJoinCommand(buffer *Buffer, args []string) tea.Cmd {
	if len(args) < 1 || len(args) > 2 {
		m.addCommandError(buffer, "usage: /join <channel> [key]")
		return nil
	}

	key := ""
	if len(args) == 2 {
		key = args[1]
	}
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}
	server, channel := buffer.Server, args[0]
	return func() tea.Msg {
		manager.Join(server, channel, key)
		return nil
	}
}

func (m *Model) handleLeaveCommand(buffer *Buffer, args []string) tea.Cmd {
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}
	server, channel := buffer.Server, buffer.Buffer
	reason := strings.Join(args, " ")
	return func() tea.Msg {
		manager.Part(server, channel, reason)
		return nil
	}
}

func (m *Model) handleCloseCommand(buffer *Buffer, args []string) tea.Cmd {
	if len(args) != 0 {
		m.addCommandError(buffer, "usage: /close")
		return nil
	}
	if buffer.Buffer == "" || irc.IsChannel(buffer.Buffer) {
		m.addCommandError(buffer, "error: /close is only available in a private message")
		return nil
	}

	return m.removeBuffer(buffer, true)
}

func (m *Model) handleListCommand(buffer *Buffer, args []string) tea.Cmd {
	if len(args) > 1 {
		m.addCommandError(buffer, "usage: /list [channel]")
		return nil
	}

	channel := ""
	if len(args) == 1 {
		channel = args[0]
	}
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}
	server := buffer.Server
	return func() tea.Msg {
		manager.List(server, channel)
		return nil
	}
}

func (m *Model) handleMsgCommand(buffer *Buffer, args []string) tea.Cmd {
	if len(args) < 2 {
		m.addCommandError(buffer, "usage: /msg <user> <message>")
		return nil
	}

	target := args[0]
	message := strings.Join(args[1:], " ")
	privateBuffer, createCmd := m.getOrCreateBuffer(buffer.Server, target)
	if privateBuffer == nil {
		return nil
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
	return m.sendMessageCmd(buffer.Server, target, message)
}

func (m *Model) sendMessageCmd(server, target, message string) tea.Cmd {
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}

	return func() tea.Msg {
		manager.Send(server, target, message)
		return nil
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
