package irc

import (
	"fmt"
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

	// For PART events, girc may not have updated its state yet
	// Schedule a refresh using WHO which will trigger RPL_ENDOFWHO with updated data
	if e.Command == girc.PART {
		client.Cmd.Who(channelName)
		return
	}

	// For NICK events, update all channels the user is in
	if e.Command == girc.NICK {
		newNick := e.Params[0]
		user := client.LookupUser(newNick)
		if user != nil {
			for _, ch := range user.ChannelList {
				c.refreshUserList(client, ch)
			}
		}
		return
	}

	c.refreshUserList(client, channelName)
}

func (c *Client) refreshUserList(client *girc.Client, channelName string) {
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

	// Use server timestamp if available (ZNC/IRCv3 server-time), fallback to current time
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	msgID, _ := e.Tags.Get("msgid")
	c.program.Send(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    target,
		Timestamp: ts,
		From:      e.Source.Name,
		Text:      e.Last(),
		MsgID:     msgID,
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

	// Use server timestamp if available, fallback to current time
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	c.program.Send(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: ts,
		From:      c.serverName,
		Text:      e.Last(),
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
	userName := e.Source.Name
	user := client.LookupUser(userName)
	if user == nil {
		return
	}

	reason := e.Last()
	message := fmt.Sprintf("%s has quit", userName)
	if reason != "" {
		message = fmt.Sprintf("%s (%s)", message, reason)
	}

	// send quit message and trigger user list refresh only for channels the user was in
	for _, channelName := range user.ChannelList {
		c.program.Send(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "--",
			Text:      message,
		})
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
				Server:    c.serverName,
				Buffer:    channelName,
				Timestamp: time.Now(),
				From:      "--",
				Text:      fmt.Sprintf("%s has joined", userName),
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
				Server:    c.serverName,
				Buffer:    channelName,
				Timestamp: time.Now(),
				From:      "--",
				Text:      fmt.Sprintf("%s has left", userName),
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

	// Sync the nickname from the server right after the connection.
	// When using ZNC, the nickname from the config file might be different from the
	// nickname configured in the ZNC, so we fetch it from there and update it
	c.program.Send(NickUpdateMsg{
		Server: c.serverName,
		Nick:   client.GetNick(),
	})

	c.program.Send(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: connected",
	})

	for _, channelName := range client.ChannelList() {
		c.program.Send(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "--",
			Text:      "irc: connected",
		})
	}
}

func (c *Client) onDisconnect(client *girc.Client, _ girc.Event) {
	c.program.Send(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: disconnected from server",
	})

	for _, channelName := range client.ChannelList() {
		c.program.Send(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "--",
			Text:      "irc: disconnected from server",
		})
	}
}

func (c *Client) onJoinError(_ *girc.Client, e girc.Event) {
	c.program.Send(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    e.Params[1],
		Timestamp: time.Now(),
		From:      "--",
		Text:      fmt.Sprintf("irc: %s", e.Last()),
	})
}

// onEchoMessage handles echo-message capability (user's own messages echoed back by the server)
func (c *Client) onEchoMessage(client *girc.Client, e girc.Event) {
	if !e.Echo || (e.Command != girc.PRIVMSG && e.Command != girc.NOTICE) {
		return
	}

	if len(e.Params) == 0 {
		return
	}
	target := e.Params[0]

	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	// This filter aims to support displaying the user's own messages in the
	// ZNC playback history correctly and support multi-client connections,
	// e.g., sending a message in client X should also display in client Y
	// if connected to both.
	// Check if this is an echo of a message we sent from this client
	// If so, mark it so the UI skips display but still stores it in history
	ownEcho := false
	if e.Source.Name == client.GetNick() {
		key := c.serverName + ":" + target + ":" + strings.TrimSpace(e.Last())
		if _, pending := c.pendingMessages.LoadAndDelete(key); pending {
			ownEcho = true
		}
		// Not in pending = sent from another client, display it
	}

	msgID, _ := e.Tags.Get("msgid")
	c.program.Send(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    target,
		Timestamp: ts,
		From:      e.Source.Name,
		Text:      e.Last(),
		MsgID:     msgID,
		OwnEcho:   ownEcho,
	})
}
