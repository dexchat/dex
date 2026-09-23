package irc

import (
	"context"
	"crypto/tls"
	"sort"
	"sync"

	"github.com/lrstanley/girc"
	"github.com/dexchat/dex/internal/config"
)

// IsChannel reports whether name is an IRC channel name (as opposed to a
// nickname/query target), keeping the girc dependency out of other packages.
func IsChannel(name string) bool {
	return girc.IsValidChannel(name)
}

// Client is for a single server connection
type Client struct {
	*girc.Client
	serverName string
	channels   []string

	// events is the single ordered path from IRC handlers to the UI. The UI
	// pulls from it with Next, so handlers never wait on Bubble Tea.
	events eventQueue

	// pendingMessages tracks messages sent from dexchat so we can ignore the echo returned by
	// the server. Since we display the message sent instantly in the UI (before sending to the server),
	// when the server echoes them back, we skip/ignore the echo to avoid duplicates.
	pendingMessages sync.Map // Key format: "server:channel:message"

	// userChannels stores the channels each user is part of to handle QUIT properly.
	// girc can remove a quitting user from its state before our handler reads it,
	// so this preserves the channel list needed to route quit messages and refreshes
	userChannelsMu sync.Mutex
	userChannels   map[string]map[string]struct{}

	// userListRefreshPending tracks pending user list channel refreshes and sends
	// them in one short batch instead of updating the UI for every IRC event
	userListRefreshMu        sync.Mutex
	userListRefreshPending   map[string]string
	userListRefreshScheduled bool
}

func NewClient(serverName string, config *config.Server) *Client {
	gircConfig := girc.Config{
		Server:     config.Address,
		Port:       config.Port,
		Nick:       config.ConnectionNickname(),
		User:       config.ConnectionUsername(),
		Name:       config.Realname,
		ServerPass: config.ConnectionPassword(),
		SSL:        config.UseSSL(),
		AllowFlood: true,
		// girc sends WHO/MODE for every self-joined channel by default.
		// On channels with a lot of users, that burst can fill ZNC's queue
		// and delay PRIVMSG traffic. We disable those extra sync requests
		// and build the displayed user list from girc's tracked state
		DisableAutoWHOOnJoin:  true,
		DisableAutoMODEOnJoin: true,
		SupportedCaps: map[string][]string{
			// echo-message is enabled to support ZNC users;
			// ZNC, by default, only records messages it receives from the IRC server.
			// Not using this capability would prevent the own user messages from being
			// preserved in ZNC playback
			"echo-message": {},
			// sever-time is enabled to fetch the correct timestamp in the ZNC playback;
			// Not using this capability would make the playback messages
			// to display with the current timestamp
			"server-time": {},
		},
	}

	if config.UseSSL() {
		gircConfig.TLSConfig = tlsConfig(config)
	}

	client := girc.New(gircConfig)
	c := &Client{
		Client:     client,
		serverName: serverName,
		channels:   config.Channels,

		userChannels: make(map[string]map[string]struct{}),
	}
	c.addHandlers()

	return c
}

// tlsConfig verifies server certificates unless the server explicitly opts
// out with ssl_skip_verify, e.g. for a self-hosted ZNC with a self-signed
// certificate.
func tlsConfig(server *config.Server) *tls.Config {
	return &tls.Config{
		ServerName:         server.Address,
		InsecureSkipVerify: server.SSLSkipVerify,
	}
}

func (c *Client) setChannelUsers(channelName string, users []string) {
	c.userChannelsMu.Lock()
	defer c.userChannelsMu.Unlock()

	for user, channels := range c.userChannels {
		delete(channels, channelName)
		if len(channels) == 0 {
			delete(c.userChannels, user)
		}
	}

	for _, user := range users {
		userID := girc.ToRFC1459(user)
		if c.userChannels[userID] == nil {
			c.userChannels[userID] = make(map[string]struct{})
		}
		c.userChannels[userID][channelName] = struct{}{}
	}
}

func (c *Client) channelsForUser(nick string) []string {
	userID := girc.ToRFC1459(nick)

	c.userChannelsMu.Lock()
	defer c.userChannelsMu.Unlock()

	channels := make([]string, 0, len(c.userChannels[userID]))
	for channelName := range c.userChannels[userID] {
		channels = append(channels, channelName)
	}
	sort.Strings(channels)
	return channels
}

func (c *Client) forgetUser(nick string) {
	userID := girc.ToRFC1459(nick)

	c.userChannelsMu.Lock()
	delete(c.userChannels, userID)
	c.userChannelsMu.Unlock()
}

// Next blocks until this server has events for the UI and returns every
// event queued so far, in order. Bursts are collected into one batch.
func (c *Client) Next(ctx context.Context) ([]Event, error) {
	return c.events.next(ctx, eventBatchWindow)
}

func (c *Client) addHandlers() {
	c.Handlers.Add(girc.CONNECTED, c.onConnect)
	c.Handlers.Add(girc.DISCONNECTED, c.onDisconnect)
	c.Handlers.Add(girc.JOIN, c.onJoin)
	c.Handlers.Add(girc.PART, c.onPart)
	c.Handlers.Add(girc.KICK, c.onKick)

	c.Handlers.Add(girc.PRIVMSG, c.onPrivmsg)
	c.Handlers.Add(girc.NOTICE, c.onServerMessage)
	c.Handlers.Add(girc.ALL_EVENTS, c.onEchoMessage) // Handle echo-message capability
	c.Handlers.AddBg(girc.TOPIC, c.onTopic)
	c.Handlers.AddBg(girc.QUIT, c.onQuit)
	c.Handlers.AddBg(girc.JOIN, c.onUserListChange)
	c.Handlers.AddBg(girc.PART, c.onUserListChange)
	c.Handlers.AddBg(girc.KICK, c.onUserListChange)
	c.Handlers.AddBg(girc.NICK, c.onUserListChange)
	c.Handlers.AddBg(girc.MODE, c.onUserListChange)

	c.Handlers.Add(girc.ERR_NOCHANMODES, c.onJoinError)
	c.Handlers.Add(girc.ERR_INVITEONLYCHAN, c.onJoinError)
	c.Handlers.Add(girc.ERR_RESTRICTED, c.onJoinError)
	c.Handlers.Add(girc.ERR_BANNEDFROMCHAN, c.onJoinError)
	c.Handlers.Add(girc.ERR_CHANNELISFULL, c.onJoinError)
	c.Handlers.Add(girc.ERR_BADCHANNELKEY, c.onJoinError)
	c.Handlers.Add(girc.ERR_NOSUCHCHANNEL, c.onJoinError)
	c.Handlers.Add(girc.ERR_TOOMANYCHANNELS, c.onJoinError)
	c.Handlers.Add(girc.ERR_BADCHANMASK, c.onJoinError)

	c.Handlers.Add(girc.RPL_LISTSTART, c.onListReply)
	c.Handlers.Add(girc.RPL_LIST, c.onListReply)
	c.Handlers.Add(girc.RPL_LISTEND, c.onListReply)
	c.Handlers.Add(girc.ERR_TOOMANYMATCHES, c.onListReply)

	c.Handlers.AddBg(girc.RPL_ENDOFNAMES, c.onUserListChange)
	c.Handlers.AddBg(girc.RPL_ENDOFWHO, c.onUserListChange)
	c.Handlers.AddBg(girc.RPL_TOPIC, c.onTopic)
	c.Handlers.AddBg(girc.RPL_WELCOME, c.onServerMessage)
	c.Handlers.AddBg(girc.RPL_MOTD, c.onServerMessage)
	c.Handlers.AddBg(girc.RPL_MOTDSTART, c.onServerMessage)
	c.Handlers.AddBg(girc.RPL_ENDOFMOTD, c.onServerMessage)
	c.Handlers.Add(girc.RPL_WELCOME, c.onNickUpdate)
	c.Handlers.Add(girc.NICK, c.onNickUpdate)
}
