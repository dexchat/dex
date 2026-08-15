package irc

import (
	"crypto/tls"
	"sort"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
)

// Client is for a single server connection
type Client struct {
	*girc.Client
	serverName string
	channels   []string
	program    *tea.Program

	// pendingMessages tracks messages sent from dexchat so we can ignore the echo returned by
	// the server. Since we display the message sent instantly in the UI (before sending to the server),
	// when the server echoes them back, we skip/ignore the echo to avoid duplicates.
	pendingMessages sync.Map // Key format: "server:channel:message"

	// messageQueue collects BufferNewMessageMsg and flushes them to bubbletea every 50ms,
	// reducing the number of messages in bubbletea's queue when a lot of messages are returned
	// by the server in the same second (for example, in ZNC)
	messageQueue   []BufferNewMessageMsg
	messageQueueMu sync.Mutex
	flushPending   bool

	// channelQueue decouples self-JOIN handling from Bubble Tea. Program.Send
	// blocks until the UI receives the message, so calling it from a normal
	// girc handler can stop the socket reader during a ZNC replay.
	channelQueue        []ChannelJoinedMsg
	channelQueueMu      sync.Mutex
	channelFlushPending bool

	// userChannels stores the channels each user is part of to handle QUIT properly.
	// girc can remove a quitting user from its state before our handler reads it,
	// so this preserves the channel list needed to route quit messages and refreshes
	userChannelsMu sync.Mutex
	userChannels   map[string]map[string]struct{}

	// userListRefreshPending tracks pending user list channel refreshes and sends
	// them in one short batch instead of updating bubbletea for every IRC event
	userListRefreshMu        sync.Mutex
	userListRefreshPending   map[string]string
	userListRefreshScheduled bool
}

func NewClient(serverName string, config *config.Server, teaProgram *tea.Program) *Client {
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
		DisableAutoWhoOnJoin:  true,
		DisableAutoModeOnJoin: true,
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

	// TODO: create a config option for this
	// Skip certificate verification for self-signed certs (useful for ZNC)
	if config.UseSSL() {
		gircConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := girc.New(gircConfig)
	c := &Client{
		Client:     client,
		serverName: serverName,
		channels:   config.Channels,
		program:    teaProgram,

		userChannels: make(map[string]map[string]struct{}),
	}
	c.addHandlers()

	return c
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

func (c *Client) queueMessage(msg BufferNewMessageMsg) {
	c.messageQueueMu.Lock()
	c.messageQueue = append(c.messageQueue, msg)
	if !c.flushPending {
		c.flushPending = true
		go func() {
			time.Sleep(50 * time.Millisecond)
			c.messageQueueMu.Lock()
			batch := c.messageQueue
			c.messageQueue = nil
			c.flushPending = false
			c.messageQueueMu.Unlock()
			if len(batch) > 0 {
				c.program.Send(BufferNewMessageBatchMsg(batch))
			}
		}()
	}
	c.messageQueueMu.Unlock()
}

func (c *Client) queueChannelJoined(msg ChannelJoinedMsg) {
	c.channelQueueMu.Lock()
	c.channelQueue = append(c.channelQueue, msg)
	if !c.channelFlushPending {
		c.channelFlushPending = true
		go func() {
			time.Sleep(50 * time.Millisecond)
			c.channelQueueMu.Lock()
			batch := c.channelQueue
			c.channelQueue = nil
			c.channelFlushPending = false
			c.channelQueueMu.Unlock()
			if len(batch) > 0 && c.program != nil {
				c.program.Send(ChannelJoinedBatchMsg(batch))
			}
		}()
	}
	c.channelQueueMu.Unlock()
}

func (c *Client) addHandlers() {
	c.Handlers.Add(girc.CONNECTED, c.onConnect)
	c.Handlers.Add(girc.DISCONNECTED, c.onDisconnect)
	c.Handlers.Add(girc.JOIN, c.onJoin)
	c.Handlers.Add(girc.PART, c.onPart)

	c.Handlers.Add(girc.PRIVMSG, c.onPrivmsg)
	c.Handlers.Add(girc.NOTICE, c.onServerMessage)
	c.Handlers.Add(girc.ALL_EVENTS, c.onEchoMessage) // Handle echo-message capability
	c.Handlers.AddBg(girc.TOPIC, c.onTopic)
	c.Handlers.AddBg(girc.QUIT, c.onQuit)
	c.Handlers.AddBg(girc.JOIN, c.onUserListChange)
	c.Handlers.AddBg(girc.PART, c.onUserListChange)
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
