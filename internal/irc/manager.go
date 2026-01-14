package irc

import (
	"fmt"
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

func (m *ClientManager) Send(server, channel, message string) error {
	if client, ok := m.clients[server]; ok {
		if !client.IsConnected() {
			return fmt.Errorf("not connected to %s", server)
		}
		if girc.IsValidChannel(channel) && !client.IsInChannel(channel) {
			return fmt.Errorf("not in channel %s", channel)
		}
		client.Cmd.Message(channel, message)
		return nil
	}
	return fmt.Errorf("server %s not found", server)
}

func (m *ClientManager) DisconnectAll() {
	for _, client := range m.clients {
		client.Close()
	}
}
