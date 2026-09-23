package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/components/channels"
	"github.com/dexchat/dex/internal/ui/components/users"
)

// ircEventsMsg carries one batch of events pulled from a server's queue.
type ircEventsMsg struct {
	server string
	events []irc.Event
}

// ircContinueMsg resumes applying events already pulled from a server.
type ircContinueMsg struct {
	server string
}

// waitIRCEvents pulls the next batch of events for server. At most one pull
// per server is outstanding, and the next one starts only after the previous
// batch has been fully applied. This keeps events in IRC order even when a
// large batch is split across several updates.
func (m *Model) waitIRCEvents(server string) tea.Cmd {
	manager := m.ircClientManager
	if manager == nil {
		return nil
	}
	return func() tea.Msg {
		events, err := manager.Next(server)
		if err != nil {
			return nil
		}
		return ircEventsMsg{server: server, events: events}
	}
}

// applyPendingIRCEvents applies up to maxPlaybackMessagesPerUpdate pulled
// events for server. It resumes on a later update while events remain and
// pulls the next batch once none are left.
func (m *Model) applyPendingIRCEvents(server string) tea.Cmd {
	pending := m.ircPending[server]
	count := min(len(pending), maxPlaybackMessagesPerUpdate)

	cmds := make([]tea.Cmd, 0, count+1)
	for _, event := range pending[:count] {
		cmds = append(cmds, m.applyIRCEvent(event))
	}

	if remaining := pending[count:]; len(remaining) > 0 {
		m.ircPending[server] = remaining
		cmds = append(cmds, func() tea.Msg { return ircContinueMsg{server: server} })
	} else {
		delete(m.ircPending, server)
		cmds = append(cmds, m.waitIRCEvents(server))
	}
	return tea.Batch(cmds...)
}

func (m *Model) applyIRCEvent(event irc.Event) tea.Cmd {
	switch event := event.(type) {
	case irc.ChannelJoinedMsg:
		_, cmd := m.getOrCreateBuffer(event.Server, event.Channel)
		return cmd

	case irc.ChannelPartedMsg:
		return m.removeChannelBuffer(event.Server, event.Channel)

	case irc.UserListMsg:
		return m.applyUserList(event)

	case irc.BufferNewMessageMsg:
		cmds := []tea.Cmd{m.processIncomingMessage(event)}
		if event.Type == irc.MessageTypeConnected && event.Buffer == "" {
			cmds = append(cmds, m.restoreDirectMessages(event.Server)...)
		}
		cmds = append(cmds, m.scheduleFlush())
		return tea.Batch(cmds...)

	case irc.ChannelTopicMsg:
		buf, cmd := m.getOrCreateBuffer(event.Server, event.Channel)
		if buf != nil {
			buf.Chat.SetTopic(event.Topic)
			// In case the channel topic is more than one line, resize only the
			// visible buffer. Inactive buffers are resized when selected.
			if buf.Key == m.activeBuffer {
				buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
			}
		}
		return cmd

	case irc.NickUpdateMsg:
		for _, buf := range m.buffers {
			if buf.Server == event.Server {
				buf.Chat.SetNickname(event.Nick)
			}
		}
		if buf := m.getActiveBuffer(); buf.Server == event.Server {
			buf.Chat.FlushQueue()
		}
		return nil

	case irc.ChannelNameUpdateMsg:
		var cmd tea.Cmd
		m.channels, cmd = m.channels.Update(channels.ChannelNameUpdateMsg{
			Server:        event.Server,
			CanonicalName: event.CanonicalName,
		})
		return cmd
	}
	return nil
}

// applyUserList updates an existing channel's members. It never creates a
// buffer: the self-JOIN always comes first, and a list that arrives after a
// self-PART must not bring the channel back.
func (m *Model) applyUserList(event irc.UserListMsg) tea.Cmd {
	buf := m.buffers[makeBufferKey(event.Server, event.Channel)]
	if buf == nil {
		return nil
	}

	buf.members = append(buf.members[:0], event.Users...)
	buf.memberPrefixes = event.Prefixes
	if buf.Key != m.activeBuffer {
		// Buffer not visible: invalidate now, render when active.
		buf.Chat.InvalidateContent()
		return nil
	}

	var cmd tea.Cmd
	buf.Users, cmd = buf.Users.Update(users.UserListMsg{Users: buf.members, Prefixes: buf.memberPrefixes})
	// We refresh chat on UserListMsg to dim nick if a user
	// sends a message then leaves channel.
	buf.Chat.RefreshContent()
	return cmd
}
