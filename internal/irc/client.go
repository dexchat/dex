package irc

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
)

// Client is for a single server connection
type Client struct {
	*girc.Client
	ServerName string
	Channels   []string
	program    *tea.Program
}

func NewClient(serverName string, config *config.Server, teaProgram *tea.Program) *Client {
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
		program:    teaProgram,
	}
}

func (c *Client) Connect() error {
	// Configure auto-join on connect
	c.Handlers.Add(girc.CONNECTED, func(client *girc.Client, e girc.Event) {
		c.program.Send(ChannelTopicMsg{
			Server:  c.ServerName,
			Channel: "",
			Topic:   "IRC: " + c.Server(),
		})

		for _, channel := range c.Channels {
			client.Cmd.Join(channel)
		}

		c.program.Send(NickUpdateMsg{
			Server: c.ServerName,
			Nick:   client.GetNick(),
		})
	})

	// TODO: handle QUIT
	// TODO: get the correct channel names (with the correct case)
	c.Handlers.Add(girc.RPL_ENDOFNAMES, c.onUserListChange)
	c.Handlers.Add(girc.RPL_ENDOFWHO, c.onUserListChange)
	c.Handlers.Add(girc.RPL_TOPIC, c.onTopic)
	c.Handlers.Add(girc.TOPIC, c.onTopic)
	c.Handlers.Add(girc.PART, c.onUserListChange)
	c.Handlers.Add(girc.JOIN, c.onUserListChange)
	c.Handlers.Add(girc.NICK, c.onUserListChange)
	c.Handlers.Add(girc.MODE, c.onUserListChange)
	c.Handlers.Add(girc.PRIVMSG, c.onPrivmsg)
	c.Handlers.Add(girc.NOTICE, c.onServerMessage)
	c.Handlers.Add(girc.RPL_WELCOME, c.onServerMessage)
	c.Handlers.Add(girc.RPL_MOTD, c.onServerMessage)
	c.Handlers.Add(girc.RPL_MOTDSTART, c.onServerMessage)
	c.Handlers.Add(girc.RPL_ENDOFMOTD, c.onServerMessage)
	c.Handlers.Add(girc.NICK, c.onNickUpdate)
	c.Handlers.Add(girc.RPL_WELCOME, c.onNickUpdate)

	if err := c.Client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.ServerName, err)
	}

	return nil
}
