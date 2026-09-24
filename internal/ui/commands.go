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
	return ircCommand(buffer, func() error {
		return manager.Join(server, channel, key)
	})
}

func (m *Model) handleLeaveCommand(buffer *Buffer, args []string) tea.Cmd {
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}
	server, channel := buffer.Server, buffer.Buffer
	reason := strings.Join(args, " ")
	return ircCommand(buffer, func() error {
		return manager.Part(server, channel, reason)
	})
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
	return ircCommand(buffer, func() error {
		return manager.List(server, channel)
	})
}

func (m *Model) handleMsgCommand(buffer *Buffer, args []string) tea.Cmd {
	if len(args) < 2 {
		m.addCommandError(buffer, "usage: /msg <user> <message>")
		return nil
	}

	target := args[0]
	message := strings.Join(args[1:], " ")
	privateBuffer := m.getOrCreateBuffer(buffer.Server, target)
	if privateBuffer == nil {
		return nil
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
	return m.sendMessageCmd(privateBuffer, message)
}

// sendMessageCmd sends message to buffer's target. Failures are reported in
// buffer, where the message was displayed.
func (m *Model) sendMessageCmd(buffer *Buffer, message string) tea.Cmd {
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}

	server, target := buffer.Server, buffer.Buffer
	return ircCommand(buffer, func() error {
		return manager.Send(server, target, message)
	})
}

// ircCommandFailedMsg reports that an IRC command could not be sent. It
// carries the buffer that issued the command, so the error is shown where the
// user typed it even if another buffer is active by then.
type ircCommandFailedMsg struct {
	server string
	key    BufferKey
	err    error
}

// ircCommand runs send outside the update loop and reports its error to
// buffer.
func ircCommand(buffer *Buffer, send func() error) tea.Cmd {
	server, key := buffer.Server, buffer.Key
	return func() tea.Msg {
		if err := send(); err != nil {
			return ircCommandFailedMsg{server: server, key: key, err: err}
		}
		return nil
	}
}

func (m *Model) showIRCCommandError(msg ircCommandFailedMsg) {
	buf := m.buffers[msg.key]
	if buf == nil {
		// The buffer was closed while the command ran.
		buf = m.buffers[makeBufferKey(msg.server, "")]
	}
	if buf != nil {
		m.addCommandError(buf, msg.err.Error())
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
