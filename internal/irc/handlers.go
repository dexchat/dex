package irc

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lrstanley/girc"
)

const userListRefreshDelay = 25 * time.Millisecond

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
		if len(e.Params) == 0 {
			return
		}
		channelName = e.Params[0]
	}

	// For NICK events, update all channels the user is in
	if e.Command == girc.NICK {
		c.scheduleUserListRefresh(client, client.ChannelList()...)
		return
	}

	// In case of user MODE events, for example
	if !girc.IsValidChannel(channelName) {
		return
	}

	c.scheduleUserListRefresh(client, channelName)
}

func (c *Client) scheduleUserListRefresh(client *girc.Client, channelNames ...string) {
	c.userListRefreshMu.Lock()
	if c.userListRefreshPending == nil {
		c.userListRefreshPending = make(map[string]string)
	}
	for _, channelName := range channelNames {
		if !girc.IsValidChannel(channelName) {
			continue
		}
		c.userListRefreshPending[girc.ToRFC1459(channelName)] = channelName
	}
	if c.userListRefreshScheduled {
		c.userListRefreshMu.Unlock()
		return
	}
	c.userListRefreshScheduled = true
	c.userListRefreshMu.Unlock()

	go func() {
		time.Sleep(userListRefreshDelay)

		c.userListRefreshMu.Lock()
		channels := make([]string, 0, len(c.userListRefreshPending))
		for _, channelName := range c.userListRefreshPending {
			channels = append(channels, channelName)
		}
		c.userListRefreshPending = nil
		c.userListRefreshScheduled = false
		c.userListRefreshMu.Unlock()

		sort.Strings(channels)
		for _, channelName := range channels {
			c.refreshUserList(client, channelName)
		}
	}()
}

func (c *Client) refreshUserList(client *girc.Client, channelName string) {
	channel := client.LookupChannel(channelName)
	if channel == nil {
		return
	}

	userList := make([]string, len(channel.UserList))

	for i, nick := range channel.UserList {
		prefix := ""
		displayNick := nick
		user := client.LookupUser(nick)

		if user != nil {
			displayNick = user.Nick
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

		userList[i] = prefix + displayNick
	}

	sortUserList(userList)
	c.setChannelUsers(channelName, userList)

	c.program.Send(UserListMsg{
		Server:  c.serverName,
		Channel: channelName,
		Users:   userList,
	})
}

func (c *Client) onPrivmsg(client *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	if e.Source == nil {
		return
	}

	target := e.Params[0]
	directMessage := !girc.IsValidChannel(target)

	// For PMs, use sender nick as the buffer identifier
	if directMessage {
		target = e.Source.Name
	}

	// Use server timestamp if available (ZNC/IRCv3 server-time), fallback to current time
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	ownEcho := false
	if e.Source.ID() == client.GetID() {
		if _, pending := c.pendingMessages.LoadAndDelete(pendingMessageKey(c.serverName, target, e.Last())); pending {
			ownEcho = true
		}
	}

	msgID, _ := e.Tags.Get("msgid")
	c.queueMessage(BufferNewMessageMsg{
		Server:        c.serverName,
		Buffer:        target,
		DirectMessage: directMessage,
		Timestamp:     ts,
		From:          e.Source.Name,
		Text:          e.Last(),
		MsgID:         msgID,
		OwnEcho:       ownEcho,
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

	c.queueMessage(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: ts,
		From:      c.serverName,
		Text:      e.Last(),
		Type:      MessageTypeServer,
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
	if e.Source == nil {
		return
	}

	userName := e.Source.Name
	user := client.LookupUser(userName)
	channels := c.channelsForUser(userName)
	if user != nil && len(user.ChannelList) > 0 {
		channels = user.ChannelList
	}
	if len(channels) == 0 {
		return
	}
	c.forgetUser(userName)

	reason := e.Last()
	host := e.Source.Host
	ident := e.Source.Ident
	message := fmt.Sprintf("%s (%s:%s) has quit", host, ident, userName)
	if reason != "" {
		message = fmt.Sprintf("%s (%s:%s) has quit (%s)", host, ident, userName, reason)
	}

	// send quit message and trigger user list refresh only for channels the user was in
	for _, channelName := range channels {
		c.queueMessage(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "<--",
			Text:      message,
		})
	}
	c.scheduleUserListRefresh(client, channels...)
}

func (c *Client) onJoin(client *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	channelName := e.Params[0]
	c.updateChannelCase(channelName)

	if e.Source != nil {
		userName := e.Source.Name
		if userName == client.GetNick() {
			c.queueChannelJoined(ChannelJoinedMsg{
				Server:  c.serverName,
				Channel: channelName,
			})
		} else {
			c.queueMessage(BufferNewMessageMsg{
				Server:    c.serverName,
				Buffer:    channelName,
				Timestamp: time.Now(),
				From:      "-->",
				Text:      fmt.Sprintf("%s (%s:%s) has joined", userName, e.Source.Host, e.Source.Ident),
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
			c.queueMessage(BufferNewMessageMsg{
				Server:    c.serverName,
				Buffer:    channelName,
				Timestamp: time.Now(),
				From:      "<--",
				Text:      fmt.Sprintf("%s (%s:%s) has left", userName, e.Source.Host, e.Source.Ident),
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

	c.queueMessage(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: connected",
		Type:      MessageTypeConnected,
	})

	for _, channelName := range client.ChannelList() {
		c.queueMessage(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "--",
			Text:      "irc: connected",
			Type:      MessageTypeConnected,
		})
	}
}

func (c *Client) onDisconnect(client *girc.Client, _ girc.Event) {
	c.queueMessage(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: disconnected from server",
		Type:      MessageTypeDisconnected,
	})

	for _, channelName := range client.ChannelList() {
		c.queueMessage(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "--",
			Text:      "irc: disconnected from server",
			Type:      MessageTypeDisconnected,
		})
	}
}

func (c *Client) onJoinError(_ *girc.Client, e girc.Event) {
	c.queueMessage(BufferNewMessageMsg{
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
	directMessage := !girc.IsValidChannel(target)

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
	if e.Source.ID() == client.GetID() {
		key := pendingMessageKey(c.serverName, target, e.Last())
		if _, pending := c.pendingMessages.LoadAndDelete(key); pending {
			ownEcho = true
		}
		// Not in pending = sent from another client, display it
	}

	msgID, _ := e.Tags.Get("msgid")
	c.queueMessage(BufferNewMessageMsg{
		Server:        c.serverName,
		Buffer:        target,
		DirectMessage: directMessage,
		Timestamp:     ts,
		From:          e.Source.Name,
		Text:          e.Last(),
		MsgID:         msgID,
		OwnEcho:       ownEcho,
	})
}
