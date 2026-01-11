package irc

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
					Server:  name,
					Channel: "",
					Time:    time.Now().Format("15:04"),
					From:    name,
					Text:    err.Error(),
				})
			}
		}(name, client)
	}
}

func (m *ClientManager) Send(server, channel, message string) {
	if client, ok := m.clients[server]; ok {
		client.Cmd.Message(channel, message)
	}
}
