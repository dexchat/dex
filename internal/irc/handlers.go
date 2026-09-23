package irc

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lrstanley/girc"
)

const userListRefreshDelay = 25 * time.Millisecond

// onEvent is the only handler dex registers with girc. For each event girc
// runs ALL_EVENTS handlers to completion before its own state tracking, and
// events are handled one at a time. Handlers called from here therefore see
// the state from just before the event, such as the channels of a quitting
// user, and queue UI events in the order the server sent them.
//
// CONNECTED is the exception: girc dispatches it from its own goroutine a few
// seconds after RPL_WELCOME, so it can interleave with other events.
func (c *Client) onEvent(client *girc.Client, e girc.Event) {
	if e.Echo {
		c.onEchoMessage(client, e)
		return
	}

	switch e.Command {
	case girc.CONNECTED:
		c.onConnect(client, e)
	case girc.DISCONNECTED:
		c.onDisconnect(client, e)

	case girc.PRIVMSG:
		c.onPrivmsg(client, e)
	case girc.NOTICE:
		c.onServerMessage(client, e)

	case girc.JOIN:
		c.onJoin(client, e)
		c.onUserListChange(client, e)
	case girc.PART:
		c.onPart(client, e)
		c.onUserListChange(client, e)
	case girc.KICK:
		c.onKick(client, e)
		c.onUserListChange(client, e)
	case girc.QUIT:
		c.onQuit(client, e)
	case girc.NICK:
		c.onNickUpdate(client, e)
		c.onUserListChange(client, e)
	case girc.MODE, girc.RPL_ENDOFNAMES, girc.RPL_ENDOFWHO:
		c.onUserListChange(client, e)

	case girc.TOPIC, girc.RPL_TOPIC:
		c.onTopic(client, e)

	case girc.RPL_WELCOME:
		c.onServerMessage(client, e)
		c.onNickUpdate(client, e)
	case girc.RPL_MOTDSTART, girc.RPL_MOTD, girc.RPL_ENDOFMOTD:
		c.onServerMessage(client, e)

	case girc.ERR_NOCHANMODES, girc.ERR_INVITEONLYCHAN, girc.ERR_RESTRICTED,
		girc.ERR_BANNEDFROMCHAN, girc.ERR_CHANNELISFULL, girc.ERR_BADCHANNELKEY,
		girc.ERR_NOSUCHCHANNEL, girc.ERR_TOOMANYCHANNELS, girc.ERR_BADCHANMASK:
		c.onJoinError(client, e)

	case girc.RPL_LISTSTART, girc.RPL_LIST, girc.RPL_LISTEND, girc.ERR_TOOMANYMATCHES:
		c.onListReply(client, e)
	}
}

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
				case perms.Prefixes != "":
					prefix = perms.Prefixes[:1]
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

	prefixOrder := girc.DefaultPrefixes
	if advertised, ok := client.GetServerOption("PREFIX"); ok {
		prefixOrder = advertised
	}
	sortUserList(userList, prefixOrder)

	c.events.push(UserListMsg{
		Server:   c.serverName,
		Channel:  channelName,
		Users:    userList,
		Prefixes: prefixSymbols(prefixOrder),
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
	isSelf := e.Source.ID() == client.GetID()

	// For PMs, use sender nick as the buffer identifier
	if directMessage && !isSelf {
		target = e.Source.Name
	}

	// Use server timestamp if available (ZNC/IRCv3 server-time), fallback to current time
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	ownEcho := false
	if isSelf {
		if _, pending := c.pendingMessages.LoadAndDelete(pendingMessageKey(c.serverName, target, e.Last())); pending {
			ownEcho = true
		}
	}

	msgID, _ := e.Tags.Get("msgid")
	c.events.push(BufferNewMessageMsg{
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

	c.events.push(ChannelTopicMsg{
		Server:  c.serverName,
		Channel: channelName,
		Topic:   e.Last(),
	})
}

func (c *Client) onServerMessage(client *girc.Client, e girc.Event) {
	if e.Command == girc.NOTICE && len(e.Params) > 0 && girc.IsValidChannel(e.Params[0]) {
		return
	}

	// A NOTICE sent directly to us (e.g. from NickServ) is a private
	// message, not a server notice, so route it like a PM/query.
	if e.Command == girc.NOTICE && e.Source != nil && len(e.Params) > 0 && !girc.IsValidChannel(e.Params[0]) {
		c.onPrivmsg(client, e)
		return
	}

	// Use server timestamp if available, fallback to current time
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: ts,
		From:      c.serverName,
		Text:      e.Last(),
		Type:      MessageTypeServer,
	})
}

// onNickUpdate reports our nickname from RPL_WELCOME or our own NICK. It runs
// before girc updates its state, so the new nickname comes from the event.
func (c *Client) onNickUpdate(client *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	switch e.Command {
	case girc.RPL_WELCOME:
		// RPL_WELCOME (001): <nick> :<welcome message>
	case girc.NICK:
		// NICK: <new nick>, sent from our current nickname.
		if e.Source == nil || e.Source.ID() != client.GetID() {
			return
		}
	default:
		return
	}

	c.events.push(NickUpdateMsg{
		Server: c.serverName,
		Nick:   e.Params[0],
	})
}

func (c *Client) onQuit(client *girc.Client, e girc.Event) {
	if e.Source == nil {
		return
	}

	// girc removes the user only after onEvent returns, so its state still
	// lists the channels they are leaving.
	userName := e.Source.Name
	user := client.LookupUser(userName)
	if user == nil || len(user.ChannelList) == 0 {
		return
	}
	channels := user.ChannelList

	reason := e.Last()
	host := e.Source.Host
	ident := e.Source.Ident
	message := fmt.Sprintf("%s (%s:%s) has quit", userName, host, ident)
	if reason != "" {
		message = fmt.Sprintf("%s (%s:%s) has quit (%s)", userName, host, ident, reason)
	}

	// send quit message and trigger user list refresh only for channels the user was in
	for _, channelName := range channels {
		c.events.push(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "<--",
			Text:      message,
			UserEvent: true,
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
			c.events.push(ChannelJoinedMsg{
				Server:  c.serverName,
				Channel: channelName,
			})
		} else {
			c.events.push(BufferNewMessageMsg{
				Server:    c.serverName,
				Buffer:    channelName,
				Timestamp: time.Now(),
				From:      "-->",
				Text:      fmt.Sprintf("%s (%s:%s) has joined", userName, e.Source.Host, e.Source.Ident),
				UserEvent: true,
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
		if userName == client.GetNick() {
			c.events.push(ChannelPartedMsg{
				Server:  c.serverName,
				Channel: channelName,
			})
			return
		}
		c.events.push(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: time.Now(),
			From:      "<--",
			Text:      fmt.Sprintf("%s (%s:%s) has left", userName, e.Source.Host, e.Source.Ident),
			UserEvent: true,
		})
	}
}

func (c *Client) onKick(client *girc.Client, e girc.Event) {
	if len(e.Params) < 2 {
		return
	}
	channelName := e.Params[0]
	kickedNick := e.Params[1]
	reason := ""
	if len(e.Params) > 2 {
		reason = e.Params[2]
	}

	if kickedNick == client.GetNick() {
		c.events.push(ChannelPartedMsg{
			Server:  c.serverName,
			Channel: channelName,
		})
		return
	}

	kickerName := c.serverName
	if e.Source != nil {
		kickerName = e.Source.Name
	}

	message := fmt.Sprintf("%s has been kicked by %s", kickedNick, kickerName)
	if reason != "" {
		message = fmt.Sprintf("%s has been kicked by %s (%s)", kickedNick, kickerName, reason)
	}

	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    channelName,
		Timestamp: time.Now(),
		From:      "<--",
		Text:      message,
		UserEvent: true,
	})
}

// updateChannelCase fixes the channel name accordingly to how it's registered in the server
// Example: the user will join the channel #idlerpg, but the true name is #idleRPG
// It should be called every time the user joins a channel
func (c *Client) updateChannelCase(channelName string) {
	for i, ch := range c.channels {
		if strings.EqualFold(ch, channelName) && ch != channelName {
			c.channels[i] = channelName
			c.events.push(ChannelNameUpdateMsg{
				Server:        c.serverName,
				CanonicalName: channelName,
			})
			return
		}
	}
}

func sortUserList(users []string, rawPrefixOrder string) {
	prefixOrder := prefixSymbols(rawPrefixOrder)

	sort.Slice(users, func(i, j int) bool {
		getNick := func(user string) string {
			if len(user) > 0 && strings.Contains(prefixOrder, user[0:1]) {
				return user[1:]
			}
			return user
		}

		getPriority := func(user string) int {
			if len(user) > 0 {
				if priority := strings.Index(prefixOrder, user[0:1]); priority >= 0 {
					return priority
				}
			}
			return len(prefixOrder)
		}

		priorityA := getPriority(users[i])
		priorityB := getPriority(users[j])

		if priorityA != priorityB {
			return priorityA < priorityB
		}

		nickA := getNick(users[i])
		nickB := getNick(users[j])

		return strings.ToLower(nickA) < strings.ToLower(nickB)
	})
}

func prefixSymbols(raw string) string {
	if _, symbols, ok := strings.Cut(raw, ")"); ok {
		return symbols
	}
	return raw
}

func (c *Client) onConnect(client *girc.Client, _ girc.Event) {
	// Auto-join on connect
	client.Cmd.Join(c.channels...)

	c.events.push(ChannelTopicMsg{
		Server:  c.serverName,
		Channel: "",
		Topic:   "IRC: " + c.Server(),
	})

	// Sync the nickname from the server right after the connection.
	// When using ZNC, the nickname from the config file might be different from the
	// nickname configured in the ZNC, so we fetch it from there and update it
	c.events.push(NickUpdateMsg{
		Server: c.serverName,
		Nick:   client.GetNick(),
	})

	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: connected",
		Type:      MessageTypeConnected,
	})

	for _, channelName := range client.ChannelList() {
		c.events.push(BufferNewMessageMsg{
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
	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: disconnected from server",
		Type:      MessageTypeDisconnected,
	})

	for _, channelName := range client.ChannelList() {
		c.events.push(BufferNewMessageMsg{
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
	channel := ""
	if len(e.Params) > 1 {
		channel = e.Params[1]
	}
	text := fmt.Sprintf("irc: %s", e.Last())
	if channel != "" {
		text = fmt.Sprintf("irc: %s: %s", channel, e.Last())
	}
	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      text,
		Type:      MessageTypeServer,
	})
}

func (c *Client) onListReply(_ *girc.Client, e girc.Event) {
	text := e.Last()
	switch e.Command {
	case girc.RPL_LISTSTART:
		text = "Channel Users Topic"
	case girc.RPL_LIST:
		if len(e.Params) < 3 {
			return
		}
		text = fmt.Sprintf("%s (%s users)", e.Params[1], e.Params[2])
		if len(e.Params) > 3 && e.Last() != "" {
			text += " " + e.Last()
		}
	case girc.RPL_LISTEND:
		if text == "" {
			text = "End of /LIST"
		}
	case girc.ERR_TOOMANYMATCHES:
		text = fmt.Sprintf("irc: %s", e.Last())
	}

	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: ts,
		From:      "--",
		Text:      text,
		Type:      MessageTypeServer,
	})
}

// onEchoMessage handles echo-message capability (user's own messages echoed back by the server)
func (c *Client) onEchoMessage(client *girc.Client, e girc.Event) {
	if e.Command != girc.PRIVMSG && e.Command != girc.NOTICE {
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
	c.events.push(BufferNewMessageMsg{
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
