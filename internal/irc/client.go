package irc

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/ui/components/users"
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
		for _, channel := range c.Channels {
			client.Cmd.Join(channel)
		}
	})

	c.Handlers.Add(girc.RPL_ENDOFNAMES, c.onUserListChange)
	c.Handlers.Add(girc.PART, c.onUserListChange)
	c.Handlers.Add(girc.JOIN, c.onUserListChange)
	c.Handlers.Add(girc.NICK, c.onUserListChange)
	c.Handlers.Add(girc.QUIT, c.onUserListChange)

	// Connect
	if err := c.Client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.ServerName, err)
	}

	return nil
}

func (c *Client) onUserListChange(client *girc.Client, e girc.Event) {
	var channelName string
	if e.Command == girc.RPL_ENDOFNAMES {
		channelName = e.Params[1]
	} else {
		channelName = e.Params[0]
	}
	log.Printf("Handler triggered for event: %q, Source: %s, Params: %v", e.Command, e.Source, e.Params)
	channel := client.LookupChannel(channelName)
	if channel == nil {
		return
	}

	userList := make([]string, len(channel.UserList))

	for i, nick := range channel.UserList {
		prefix := ""
		user := client.LookupUser(nick)

		if perms, ok := user.Perms.Lookup(channelName); ok {
			switch {
			case perms.Owner:
				prefix = girc.OwnerPrefix
			case perms.Admin:
				prefix = girc.AdminPrefix
			case perms.Op:
				prefix = girc.OperatorPrefix
			case perms.HalfOp:
				prefix = girc.HalfOperatorPrefix
			case perms.Voice:
				prefix = girc.VoicePrefix
			}
		}

		userList[i] = prefix + nick
	}

	c.program.Send(users.UserListMsg(userList))
}
