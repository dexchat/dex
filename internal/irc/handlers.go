package irc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lrstanley/girc"
)

// onEvent is the only handler dex registers with girc. For each event girc
// runs ALL_EVENTS handlers to completion before its own state tracking, and
// events are handled one at a time. Handlers called from here therefore see
// the state from just before the event, such as the channels of a quitting
// user, and queue UI events in the order the server sent them.
//
// girc reports UPDATE_STATE after applying a state change, which is when
// user list changes recorded for the event are queued.
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
		c.onNickChange(client, e)
		c.onUserListChange(client, e)
	case girc.MODE, girc.RPL_ENDOFNAMES, girc.RPL_ENDOFWHO:
		c.onUserListChange(client, e)

	case girc.TOPIC, girc.RPL_TOPIC:
		c.onTopic(client, e)

	case girc.RPL_WELCOME:
		c.registered.Store(true)
		c.onServerMessage(client, e)
		c.onNickUpdate(client, e)
	case girc.RPL_MOTDSTART, girc.RPL_MOTD, girc.RPL_ENDOFMOTD:
		c.onServerMessage(client, e)

	case girc.ERR_NOCHANMODES, girc.ERR_INVITEONLYCHAN, girc.ERR_RESTRICTED,
		girc.ERR_BANNEDFROMCHAN, girc.ERR_CHANNELISFULL, girc.ERR_BADCHANNELKEY,
		girc.ERR_NOSUCHCHANNEL, girc.ERR_TOOMANYCHANNELS, girc.ERR_BADCHANMASK:
		c.onErrorReply(client, e)

	case girc.ERR_ERRONEUSNICKNAME, girc.ERR_NICKNAMEINUSE, girc.ERR_NICKCOLLISION,
		girc.ERR_UNAVAILRESOURCE, errNickTooFast:
		c.onErrorReply(client, e)

	case girc.RPL_LISTSTART, girc.RPL_LIST, girc.RPL_LISTEND, girc.ERR_TOOMANYMATCHES:
		c.onListReply(client, e)

	case girc.RPL_WHOISUSER, girc.RPL_WHOISSERVER, girc.RPL_WHOISOPERATOR,
		girc.RPL_WHOISIDLE, girc.RPL_ENDOFWHOIS, girc.RPL_WHOISCHANNELS,
		girc.RPL_AWAY, girc.RPL_WHOISACCOUNT, girc.RPL_WHOISACTUALLY,
		girc.RPL_WHOISHOST, girc.RPL_WHOISMODES, girc.RPL_WHOISCERTFP,
		girc.RPL_WHOISREGNICK, girc.RPL_WHOISSPECIAL, rplWhoisSecure,
		rplWhoisBot, girc.ERR_NOSUCHNICK:
		c.onWhoisReply(client, e)

	case girc.UPDATE_STATE:
		c.queueDeferredUserListChanges()
	}
}

// userListChanged marks a channel whose user list must be sent to the UI. It
// never leaves the client: Next replaces these markers with one snapshot per
// channel, so a burst of membership events costs one snapshot per batch.
type userListChanged struct {
	channel string
}

func (userListChanged) ircEvent() {}

func (c *Client) onUserListChange(client *girc.Client, e girc.Event) {
	switch e.Command {
	case girc.RPL_ENDOFNAMES, girc.RPL_ENDOFWHO:
		// Format: <nick> <channel> :End of /NAMES list
		// The replies before it already updated girc's state.
		if len(e.Params) >= 2 {
			c.queueUserListChanges(e.Params[1])
		}
	case girc.NICK:
		// girc renames the user after this handler returns, so the old
		// nickname still identifies the channels to update.
		if e.Source == nil {
			return
		}
		if user := client.LookupUser(e.Source.Name); user != nil {
			c.deferUserListChanges(user.ChannelList...)
		}
	default:
		// JOIN, PART, KICK, and MODE name the channel first. User MODE
		// events name a nickname instead and are skipped.
		if len(e.Params) > 0 {
			c.deferUserListChanges(e.Params[0])
		}
	}
}

func (c *Client) queueUserListChanges(channelNames ...string) {
	for _, channelName := range channelNames {
		if girc.IsValidChannel(channelName) {
			c.events.push(userListChanged{channel: channelName})
		}
	}
}

// deferUserListChanges records channels changed by the current event. girc
// applies the event only after onEvent returns; queueing the markers then
// guarantees that a snapshot taken for them includes the change.
func (c *Client) deferUserListChanges(channelNames ...string) {
	c.deferredUserListsMu.Lock()
	defer c.deferredUserListsMu.Unlock()
	c.deferredUserLists = append(c.deferredUserLists, channelNames...)
}

func (c *Client) queueDeferredUserListChanges() {
	c.deferredUserListsMu.Lock()
	channelNames := c.deferredUserLists
	c.deferredUserLists = nil
	c.deferredUserListsMu.Unlock()

	c.queueUserListChanges(channelNames...)
}

// resolveUserListChanges replaces userListChanged markers with one user list
// snapshot per channel, appended after the batch's other events. Markers are
// queued only after girc applies their change, so each snapshot includes it.
func (c *Client) resolveUserListChanges(events []Event) []Event {
	var channelNames []string
	seen := make(map[string]bool)
	resolved := events[:0]
	for _, event := range events {
		changed, ok := event.(userListChanged)
		if !ok {
			resolved = append(resolved, event)
			continue
		}
		id := girc.ToRFC1459(changed.channel)
		if !seen[id] {
			seen[id] = true
			channelNames = append(channelNames, changed.channel)
		}
	}

	for _, channelName := range channelNames {
		if snapshot, ok := c.userListSnapshot(channelName); ok {
			resolved = append(resolved, snapshot)
		}
	}
	return resolved
}

// userListSnapshot builds the displayed user list for channelName from girc's
// state. It reports false when girc no longer tracks the channel.
func (c *Client) userListSnapshot(channelName string) (UserListMsg, bool) {
	client := c.Client
	channel := client.LookupChannel(channelName)
	if channel == nil {
		return UserListMsg{}, false
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

	return UserListMsg{
		Server:   c.serverName,
		Channel:  channelName,
		Users:    userList,
		Prefixes: prefixSymbols(prefixOrder),
	}, true
}

func (c *Client) onPrivmsg(_ *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	if e.Source == nil {
		return
	}

	target := e.Params[0]
	directMessage := !girc.IsValidChannel(target)

	// For PMs, use sender nick as the buffer identifier. Messages from us
	// arrive as echo events and are handled by onEchoMessage instead.
	if directMessage {
		target = e.Source.Name
	}

	// Use server timestamp if available (ZNC/IRCv3 server-time), fallback to current time
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	text, action := decodeAction(e.Last())
	msgID, _ := e.Tags.Get("msgid")
	c.events.push(BufferNewMessageMsg{
		Server:        c.serverName,
		Buffer:        target,
		DirectMessage: directMessage,
		Timestamp:     ts,
		From:          e.Source.Name,
		Text:          text,
		Action:        action,
		MsgID:         msgID,
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

// onNickChange shows a NICK in the chat. Our own change is shown only in the
// server buffer, where /nick errors appear. Another user's change is shown
// in the channels shared with them, like QUIT; girc renames the user only
// after onEvent returns, so the old nickname still finds those channels.
func (c *Client) onNickChange(client *girc.Client, e girc.Event) {
	if e.Source == nil || len(e.Params) == 0 {
		return
	}
	newNick := e.Params[0]
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	if e.Source.ID() == client.GetID() {
		c.events.push(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    "",
			Timestamp: ts,
			From:      "--",
			Text:      "You are now known as " + newNick,
			Type:      MessageTypeServer,
		})
		return
	}

	user := client.LookupUser(e.Source.Name)
	if user == nil {
		return
	}
	text := fmt.Sprintf("%s is now known as %s", e.Source.Name, newNick)
	for _, channelName := range user.ChannelList {
		c.events.push(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    channelName,
			Timestamp: ts,
			From:      "---",
			Text:      text,
			UserEvent: true,
		})
	}
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

	// Send the quit message and refresh the user list only in the user's channels.
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
	c.deferUserListChanges(channels...)
}

func (c *Client) onJoin(client *girc.Client, e girc.Event) {
	if len(e.Params) == 0 {
		return
	}
	channelName := e.Params[0]

	if e.Source != nil {
		userName := e.Source.Name
		if userName == client.GetNick() {
			c.updateChannelCase(channelName)
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

// updateChannelCase reports the server's spelling of a configured channel.
// Example: the user will join the channel #idlerpg, but the true name is #idleRPG
// It should be called every time the user joins a channel. The update is sent
// on every such join; applying it again in the UI is harmless.
func (c *Client) updateChannelCase(channelName string) {
	for _, ch := range c.channels {
		if strings.EqualFold(ch, channelName) && ch != channelName {
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

// errNickTooFast is sent by Solanum, UnrealIRCd, and other servers when
// nickname changes are rate limited. girc has no constant for it.
const errNickTooFast = "438"

// onErrorReply shows an error numeric of the form
// "<client> <subject> :<text>" in the server buffer, where subject is the
// rejected channel or nickname.
func (c *Client) onErrorReply(_ *girc.Client, e girc.Event) {
	subject := ""
	if len(e.Params) > 2 {
		subject = e.Params[1]
	}
	text := fmt.Sprintf("irc: %s", e.Last())
	if subject != "" {
		text = fmt.Sprintf("irc: %s: %s", subject, e.Last())
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

// WHOIS numerics that girc has no constants for.
const (
	rplWhoisSecure = "671"
	rplWhoisBot    = "335"
)

// onWhoisReply shows one WHOIS reply line in the server buffer, like /list.
// Every reply names the queried nickname as its second parameter. RPL_AWAY
// and ERR_NOSUCHNICK also answer a message sent to that nickname, and read
// correctly in both cases.
func (c *Client) onWhoisReply(_ *girc.Client, e girc.Event) {
	if len(e.Params) < 2 {
		return
	}
	nick := e.Params[1]
	// Parameters between the nickname and the trailing text.
	var middle []string
	if len(e.Params) > 3 {
		middle = e.Params[2 : len(e.Params)-1]
	}

	var text string
	switch e.Command {
	case girc.RPL_WHOISUSER:
		// <me> <nick> <user> <host> * :<realname>
		if len(e.Params) < 4 {
			return
		}
		text = fmt.Sprintf("%s (%s@%s): %s", nick, e.Params[2], e.Params[3], e.Last())
	case girc.RPL_WHOISSERVER:
		// <me> <nick> <server> :<server info>
		if len(e.Params) < 3 {
			return
		}
		text = fmt.Sprintf("%s is connected to %s", nick, e.Params[2])
		if len(e.Params) > 3 && e.Last() != "" {
			text += " (" + e.Last() + ")"
		}
	case girc.RPL_WHOISIDLE:
		// <me> <nick> <seconds> [<signon>] :seconds idle, signon time
		if len(middle) < 1 {
			return
		}
		idle, err := strconv.ParseInt(middle[0], 10, 64)
		if err != nil {
			return
		}
		text = fmt.Sprintf("%s has been idle %s", nick, time.Duration(idle)*time.Second)
		if len(middle) > 1 {
			if signon, err := strconv.ParseInt(middle[1], 10, 64); err == nil {
				text += ", signed on " + time.Unix(signon, 0).Format("2006-01-02 15:04")
			}
		}
	case girc.RPL_WHOISCHANNELS:
		// <me> <nick> :{[prefix]<channel> }
		text = fmt.Sprintf("%s is on %s", nick, strings.TrimSpace(e.Last()))
	case girc.RPL_AWAY:
		// <me> <nick> :<away message>
		text = fmt.Sprintf("%s is away: %s", nick, e.Last())
	case girc.RPL_WHOISACCOUNT:
		// <me> <nick> <account> :is logged in as
		if len(middle) < 1 {
			return
		}
		text = fmt.Sprintf("%s is logged in as %s", nick, middle[0])
	case girc.RPL_ENDOFWHOIS:
		text = fmt.Sprintf("%s: %s", nick, e.Last())
	case girc.ERR_NOSUCHNICK:
		text = fmt.Sprintf("irc: %s: %s", nick, e.Last())
	default:
		// The remaining replies are "<me> <nick> [<value>...] :<text>",
		// where the text reads as a predicate of the nickname.
		text = nick + " " + e.Last()
		if len(middle) > 0 {
			text += ": " + strings.Join(middle, " ")
		}
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

// onEchoMessage handles our own PRIVMSG and NOTICE. girc marks every message
// from our nickname as an echo, whether it is an echo-message reply, bouncer
// playback, or a message sent from another client on the same account.
func (c *Client) onEchoMessage(_ *girc.Client, e girc.Event) {
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

	// A message sent from this client was shown when it was sent, so its echo
	// is only stored. Anything else, such as playback or a message sent from
	// another client, is shown.
	ownEcho := c.sent.consume(pendingMessageKey(c.serverName, target, e.Last()), time.Now())

	text, action := e.Last(), false
	if e.Command == girc.PRIVMSG {
		text, action = decodeAction(text)
	}

	msgID, _ := e.Tags.Get("msgid")
	c.events.push(BufferNewMessageMsg{
		Server:        c.serverName,
		Buffer:        target,
		DirectMessage: directMessage,
		Timestamp:     ts,
		From:          e.Source.Name,
		Text:          text,
		Action:        action,
		MsgID:         msgID,
		OwnEcho:       ownEcho,
	})
}
