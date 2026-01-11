package irc

import (
	"strings"
	"time"

	"github.com/lrstanley/girc"
)

func (c *Client) onUserListChange(client *girc.Client, e girc.Event) {
	var channelName string
	switch e.Command {
	case girc.RPL_ENDOFNAMES, girc.RPL_ENDOFWHO:
		// Format: <nick> <channel> :End of /NAMES list
		if len(e.Params) < 2 {
			return
		}
		channelName = e.Params[1]
	default:
		channelName = e.Params[0]
	}

	// In case of user MODE events, for example
	if !girc.IsValidChannel(channelName) {
		return
	}

	// For JOIN/PART events, girc may not have updated its state yet
	// Schedule a refresh using WHO which will trigger RPL_ENDOFWHO with updated data
	if e.Command == girc.JOIN || e.Command == girc.PART {
		client.Cmd.Who(channelName)
		return
	}

	channel := client.LookupChannel(channelName)
	if channel == nil {
		return
	}

	userList := make([]string, len(channel.UserList))

	for i, nick := range channel.UserList {
		prefix := ""
		user := client.LookupUser(nick)

		if user != nil {
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
		}

		userList[i] = prefix + nick
	}

	c.program.Send(UserListMsg{
		Server:  c.ServerName,
		Channel: channelName,
		Users:   userList,
	})
}

func (c *Client) onPrivmsg(_ *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}

	channelName := e.Params[0]
	// TODO: handle private messages
	if !girc.IsValidChannel(channelName) {
		return
	}

	c.program.Send(BufferNewMessageMsg{
		Server:  c.ServerName,
		Channel: channelName,
		Time:    time.Now().Format("15:04"),
		From:    e.Source.Name,
		Text:    e.Last(),
	})
}

func (c *Client) onTopic(_ *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}

	var channelName string
	if e.Command == girc.RPL_TOPIC {
		// RPL_TOPIC (332): <nick> <channelName> :<topic>
		if len(e.Params) < 2 {
			return
		}
		channelName = e.Params[1]
	} else {
		// TOPIC: <channelName> :<topic>
		channelName = e.Params[0]
	}

	c.program.Send(ChannelTopicMsg{
		Server:  c.ServerName,
		Channel: channelName,
		Topic:   e.Last(),
	})
}

func (c *Client) onServerMessage(_ *girc.Client, e girc.Event) {
	if e.Command == girc.NOTICE && len(e.Params) > 0 && girc.IsValidChannel(e.Params[0]) {
		return
	}

	c.program.Send(BufferNewMessageMsg{
		Server:  c.ServerName,
		Channel: "",
		Time:    time.Now().Format("15:04"),
		From:    c.ServerName,
		Text:    e.Last(),
	})
}

func (c *Client) onNickUpdate(client *girc.Client, e girc.Event) {
	if e.Source.Name == client.GetNick() || e.Params[0] == client.GetNick() {
		c.program.Send(NickUpdateMsg{
			Server: c.ServerName,
			Nick:   client.GetNick(),
		})
	}
}

func (c *Client) onQuit(client *girc.Client, e girc.Event) {
	// WHO triggers RPL_ENDOFWHO which calls onUserListChange
	for _, channelName := range c.Channels {
		client.Cmd.Who(channelName)
	}
}

func (c *Client) onJoin(_ *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	channelName := e.Params[0]
	c.updateChannelCase(channelName)
}

func (c *Client) updateChannelCase(channelName string) {
	for i, ch := range c.Channels {
		if strings.EqualFold(ch, channelName) && ch != channelName {
			c.Channels[i] = channelName
			c.program.Send(ChannelNameUpdateMsg{
				Server:        c.ServerName,
				CanonicalName: channelName,
			})
			return
		}
	}
}
