package irc

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
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
			if err := c.Connect(); err != nil {
				c.program.Send(BufferNewMessageMsg{
					Server: name,
					Buffer: "",
					Time:   time.Now().Format("15:04"),
					From:   name,
					Text:   err.Error(),
				})
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
			Server: server,
			Buffer: channel,
			Time:   time.Now().Format("15:04"),
			From:   "--",
			Text:   text,
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
	key := server + ":" + channel + ":" + strings.TrimSpace(message)
	client.pendingMessages.Store(key, struct{}{})

	client.Cmd.Message(channel, message)
}

func (m *ClientManager) DisconnectAll() {
	for _, client := range m.clients {
		client.Close()
	}
}
