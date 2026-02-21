package irc

import (
	"crypto/tls"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	}
	c.addHandlers()

	return c
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

func (c *Client) addHandlers() {
	c.Handlers.Add(girc.CONNECTED, c.onConnect)
	c.Handlers.Add(girc.DISCONNECTED, c.onDisconnect)
	c.Handlers.Add(girc.JOIN, c.onJoin)
	c.Handlers.Add(girc.PART, c.onPart)

	c.Handlers.AddBg(girc.PRIVMSG, c.onPrivmsg)
	c.Handlers.AddBg(girc.NOTICE, c.onServerMessage)
	c.Handlers.AddBg(girc.ALL_EVENTS, c.onEchoMessage) // Handle echo-message capability
	c.Handlers.AddBg(girc.TOPIC, c.onTopic)
	c.Handlers.AddBg(girc.QUIT, c.onQuit)
	c.Handlers.AddBg(girc.PART, c.onUserListChange)
	c.Handlers.AddBg(girc.NICK, c.onUserListChange)
	c.Handlers.AddBg(girc.MODE, c.onUserListChange)

	c.Handlers.Add(girc.ERR_NOCHANMODES, c.onJoinError)
	c.Handlers.Add(girc.ERR_INVITEONLYCHAN, c.onJoinError)
	c.Handlers.Add(girc.ERR_RESTRICTED, c.onJoinError)
	c.Handlers.Add(girc.ERR_BANNEDFROMCHAN, c.onJoinError)
	c.Handlers.Add(girc.ERR_CHANNELISFULL, c.onJoinError)
	c.Handlers.Add(girc.ERR_BADCHANNELKEY, c.onJoinError)

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
