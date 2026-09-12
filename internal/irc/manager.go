package irc

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/girc"
	"github.com/dexchat/dex/internal/config"
)

// ClientManager handles multiple clients/server connections
// With more than one connection, the UI wouldn't know which channel to send a message
type ClientManager struct {
	clients map[string]*Client
}

func NewClientManager(servers []*config.Server, p *tea.Program) *ClientManager {
	m := &ClientManager{clients: make(map[string]*Client)}
	for _, server := range servers {
		m.clients[server.Name] = NewClient(server.Name, server, p)
	}
	return m
}

func (m *ClientManager) ConnectAll() {
	for name, client := range m.clients {
		go func(name string, c *Client) {
			attempt := 0
			retryDelays := []int{10, 20, 40, 80, 160, 300}

			for {
				if err := c.Connect(); err != nil {
					var backoff int
					if attempt < len(retryDelays) {
						backoff = retryDelays[attempt]
					} else {
						backoff = retryDelays[len(retryDelays)-1]
					}
					c.program.Send(BufferNewMessageMsg{
						Server:    name,
						Buffer:    "",
						Timestamp: time.Now(),
						From:      "--",
						Text:      fmt.Sprintf("irc: %v, reconnecting in %d seconds...", err, int(backoff)),
						Type:      MessageTypeDisconnected,
					})
					time.Sleep(time.Duration(backoff) * time.Second)
					attempt++
					continue
				}
				return
			}
		}(name, client)
	}
}

func (m *ClientManager) Send(server, channel, message string) {
	client, ok := m.clients[server]
	if !ok {
		return
	}

	sendError := func(text string) {
		client.program.Send(BufferNewMessageMsg{
			Server:    server,
			Buffer:    channel,
			Timestamp: time.Now(),
			From:      "--",
			Text:      text,
		})
	}

	if !client.IsConnected() {
		sendError(fmt.Sprintf("irc: not connected to %s", server))
		return
	}
	if girc.IsValidChannel(channel) && !client.IsInChannel(channel) {
		sendError(fmt.Sprintf("irc: not in channel %s", channel))
		return
	}

	// TrimSpace normalizes whitespace because IRC servers may strip or add
	// leading/trailing spaces, leading to the message being sent twice as
	// the echo will differ from what we sent
	key := pendingMessageKey(server, channel, message)
	client.pendingMessages.Store(key, struct{}{})

	client.Cmd.Message(channel, message)
}

func (m *ClientManager) Part(server, channel, reason string) {
	client, ok := m.clients[server]
	if !ok {
		return
	}

	sendError := func(text string) {
		client.program.Send(BufferNewMessageMsg{
			Server:    server,
			Buffer:    channel,
			Timestamp: time.Now(),
			From:      "--",
			Text:      text,
			Type:      MessageTypeServer,
		})
	}

	if !client.IsConnected() {
		sendError(fmt.Sprintf("irc: not connected to %s", server))
		return
	}
	if !girc.IsValidChannel(channel) {
		sendError("irc: /leave is only available in a channel")
		return
	}
	if !client.IsInChannel(channel) {
		sendError(fmt.Sprintf("irc: not in channel %s", channel))
		return
	}

	if reason == "" {
		client.Cmd.Part(channel)
		return
	}
	client.Cmd.PartMessage(channel, reason)
}

func (m *ClientManager) Join(server, channel, key string) {
	client, ok := m.clients[server]
	if !ok {
		return
	}

	sendError := func(text string) {
		client.program.Send(BufferNewMessageMsg{
			Server:    server,
			Buffer:    "",
			Timestamp: time.Now(),
			From:      "--",
			Text:      text,
			Type:      MessageTypeServer,
		})
	}

	if !client.IsConnected() {
		sendError(fmt.Sprintf("irc: not connected to %s", server))
		return
	}
	if !girc.IsValidChannel(channel) || strings.Contains(channel, ",") {
		sendError(fmt.Sprintf("irc: invalid channel %s", channel))
		return
	}
	if client.IsInChannel(channel) {
		sendError(fmt.Sprintf("irc: already in channel %s", channel))
		return
	}

	if key == "" {
		client.Cmd.Join(channel)
		return
	}
	client.Cmd.JoinKey(channel, key)
}

func (m *ClientManager) List(server, channel string) {
	client, ok := m.clients[server]
	if !ok {
		return
	}

	sendError := func(text string) {
		client.program.Send(BufferNewMessageMsg{
			Server:    server,
			Buffer:    "",
			Timestamp: time.Now(),
			From:      "--",
			Text:      text,
			Type:      MessageTypeServer,
		})
	}

	if !client.IsConnected() {
		sendError(fmt.Sprintf("irc: not connected to %s", server))
		return
	}
	if channel != "" && (!girc.IsValidChannel(channel) || strings.Contains(channel, ",")) {
		sendError(fmt.Sprintf("irc: invalid channel %s", channel))
		return
	}

	if channel == "" {
		client.Cmd.List()
		return
	}
	client.Cmd.List(channel)
}

func pendingMessageKey(server, target, message string) string {
	return strings.ToLower(server) + ":" + strings.ToLower(target) + ":" + strings.TrimSpace(message)
}

func (m *ClientManager) DisconnectAll() {
	for _, client := range m.clients {
		client.Close()
	}
}
