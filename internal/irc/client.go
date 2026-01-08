package irc

import (
	"fmt"

	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
)

// Client is for a single server connection
type Client struct {
	*girc.Client
	ServerName string
	Channels   []string
}

func NewClient(serverName string, config *config.Server) *Client {
	client := girc.New(girc.Config{
		Server: config.Address,
		Port:   config.Port,
		Nick:   config.Nickname,
		User:   config.Username,
		Name:   config.Realname,
	})

	return &Client{
		Client:     client,
		ServerName: serverName,
		Channels:   config.Channels,
	}
}

func (c *Client) Connect() error {
	// Configure auto-join on connect
	c.Handlers.Add(girc.CONNECTED, func(client *girc.Client, e girc.Event) {
		for _, channel := range c.Channels {
			client.Cmd.Join(channel)
		}
	})

	// Connect
	if err := c.Client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.ServerName, err)
	}

	return nil
}
