package irc

import (
	"fmt"
	"math"
	"sort"
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

	sortUserList(userList)

	c.program.Send(UserListMsg{
		Server:  c.serverName,
		Channel: channelName,
		Users:   userList,
	})
}

func (c *Client) onPrivmsg(_ *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}

	target := e.Params[0]

	// For PMs, use sender nick as the buffer identifier
	if !girc.IsValidChannel(target) {
		target = e.Source.Name
	}

	c.program.Send(BufferNewMessageMsg{
		Server: c.serverName,
		Buffer: target,
		Time:   time.Now().Format("15:04"),
		From:   e.Source.Name,
		Text:   e.Last(),
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
		Server:  c.serverName,
		Channel: channelName,
		Topic:   e.Last(),
	})
}

func (c *Client) onServerMessage(_ *girc.Client, e girc.Event) {
	if e.Command == girc.NOTICE && len(e.Params) > 0 && girc.IsValidChannel(e.Params[0]) {
		return
	}

	c.program.Send(BufferNewMessageMsg{
		Server: c.serverName,
		Buffer: "",
		Time:   time.Now().Format("15:04"),
		From:   c.serverName,
		Text:   e.Last(),
	})
}

func (c *Client) onNickUpdate(client *girc.Client, e girc.Event) {
	if e.Source.Name == client.GetNick() || e.Params[0] == client.GetNick() {
		c.program.Send(NickUpdateMsg{
			Server: c.serverName,
			Nick:   client.GetNick(),
		})
	}
}

func (c *Client) onQuit(client *girc.Client, e girc.Event) {
	// WHO triggers RPL_ENDOFWHO which calls onUserListChange
	for _, channelName := range c.channels {
		client.Cmd.Who(channelName)
	}
}

func (c *Client) onJoin(client *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	channelName := e.Params[0]
	c.updateChannelCase(channelName)

	if e.Source != nil {
		userName := e.Source.Name
		// do not display self-join
		if userName != client.GetNick() {
			c.program.Send(BufferNewMessageMsg{
				Server: c.serverName,
				Buffer: channelName,
				Time:   time.Now().Format("15:04"),
				From:   "--",
				Text:   fmt.Sprintf("%s has joined", userName),
			})
		}
	}
}

func (c *Client) onPart(client *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	channelName := e.Params[0]

	if e.Source != nil {
		userName := e.Source.Name
		// do not display self-part
		if userName != client.GetNick() {
			c.program.Send(BufferNewMessageMsg{
				Server: c.serverName,
				Buffer: channelName,
				Time:   time.Now().Format("15:04"),
				From:   "--",
				Text:   fmt.Sprintf("%s has left", userName),
			})
		}
	}
}

// updateChannelCase fixes the channel name accordingly to how it's registered in the server
// Example: the user will join the channel #idlerpg, but the true name is #idleRPG
// It should be called every time the user joins a channel
func (c *Client) updateChannelCase(channelName string) {
	for i, ch := range c.channels {
		if strings.EqualFold(ch, channelName) && ch != channelName {
			c.channels[i] = channelName
			c.program.Send(ChannelNameUpdateMsg{
				Server:        c.serverName,
				CanonicalName: channelName,
			})
			return
		}
	}
}

func sortUserList(users []string) {
	sort.Slice(users, func(i, j int) bool {
		getNick := func(user string) string {
			if len(user) > 0 {
				switch user[0:1] {
				case girc.OwnerPrefix, girc.AdminPrefix, girc.OperatorPrefix, girc.HalfOperatorPrefix, girc.VoicePrefix:
					return user[1:]
				}
			}
			return user
		}

		getPriority := func(user string) int {
			if len(user) > 0 {
				switch user[0:1] {
				case girc.OwnerPrefix:
					return 5
				case girc.AdminPrefix:
					return 4
				case girc.OperatorPrefix:
					return 3
				case girc.HalfOperatorPrefix:
					return 2
				case girc.VoicePrefix:
					return 1
				}
			}
			return 0
		}

		priorityA := getPriority(users[i])
		priorityB := getPriority(users[j])

		if priorityA != priorityB {
			return priorityA > priorityB
		}

		nickA := getNick(users[i])
		nickB := getNick(users[j])

		return strings.ToLower(nickA) < strings.ToLower(nickB)
	})
}

func (c *Client) onConnect(client *girc.Client, _ girc.Event) {
	// Auto-join on connect
	client.Cmd.Join(c.channels...)

	c.program.Send(ChannelTopicMsg{
		Server:  c.serverName,
		Channel: "",
		Topic:   "IRC: " + c.Server(),
	})
}

func (c *Client) startReconnectLoop() {
	go func() {
		attempt := 0
		for {
			backoffSeconds := math.Min(math.Pow(2, float64(attempt)), 300)

			c.program.Send(BufferNewMessageMsg{
				Server: c.serverName,
				Buffer: "",
				Time:   time.Now().Format("15:04"),
				From:   "--",
				Text:   fmt.Sprintf("irc: reconnecting in %d seconds...", int(backoffSeconds)),
			})

			time.Sleep(time.Duration(backoffSeconds) * time.Second)

			if err := c.Client.Connect(); err == nil {
				c.program.Send(BufferNewMessageMsg{
					Server: c.serverName,
					Buffer: "",
					Time:   time.Now().Format("15:04"),
					From:   "--",
					Text:   "irc: reconnected to server",
				})
				return
			} else {
				c.program.Send(BufferNewMessageMsg{
					Server: c.serverName,
					Buffer: "",
					Time:   time.Now().Format("15:04"),
					From:   "--",
					Text:   fmt.Sprintf("irc: reconnect failed: %v", err),
				})
			}

			attempt++
		}
	}()
}

func (c *Client) onDisconnect(_ *girc.Client, _ girc.Event) {
	c.program.Send(BufferNewMessageMsg{
		Server: c.serverName,
		Buffer: "",
		Time:   time.Now().Format("15:04"),
		From:   "--",
		Text:   "irc: disconnected from server",
	})

	for _, channelName := range c.channels {
		c.program.Send(BufferNewMessageMsg{
			Server: c.serverName,
			Buffer: channelName,
			Time:   time.Now().Format("15:04"),
			From:   "--",
			Text:   "irc: disconnected from server",
		})
	}

	c.startReconnectLoop()
}

func (c *Client) onJoinError(_ *girc.Client, e girc.Event) {
	c.program.Send(BufferNewMessageMsg{
		Server: c.serverName,
		Buffer: e.Params[1],
		Time:   time.Now().Format("15:04"),
		From:   "--",
		Text:   fmt.Sprintf("irc: %s", e.Last()),
	})
}
