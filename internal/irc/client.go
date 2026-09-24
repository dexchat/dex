package irc

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dexchat/dex/internal/config"
	"github.com/lrstanley/girc"
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

	// channels are the configured channels joined on every connection. They
	// are read from several goroutines and must not be modified.
	channels []string

	// registered reports whether the current connection received
	// RPL_WELCOME. connectLoop resets its backoff after such a connection.
	registered atomic.Bool

	// events is the single ordered path from IRC handlers to the UI. The UI
	// pulls from it with Next, so handlers never wait on Bubble Tea.
	events eventQueue

	// sent tracks messages sent from dex. The UI shows them when they are
	// sent, so their echoes are stored in history but not shown again.
	sent sentMessages

	// deferredUserLists holds channels whose user list changed in an event
	// girc has not applied yet. They are queued once girc reports
	// UPDATE_STATE, which it sends after applying the change.
	deferredUserListsMu sync.Mutex
	deferredUserLists   []string
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
	}
	client.Config.RecoverFunc = c.onHandlerPanic
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

// Next blocks until this server has events for the UI and returns every
// event queued so far, in order. Bursts are collected into one batch, and
// user lists changed by the batch are sent after its other events.
func (c *Client) Next(ctx context.Context) ([]Event, error) {
	for {
		events, err := c.events.next(ctx, eventBatchWindow)
		if err != nil {
			return nil, err
		}
		// A batch can hold only markers for channels girc no longer tracks.
		if events = c.resolveUserListChanges(events); len(events) > 0 {
			return events, nil
		}
	}
}

func (c *Client) addHandlers() {
	c.Handlers.Add(girc.ALL_EVENTS, c.onEvent)
}

// onHandlerPanic keeps a bug in an IRC handler from crashing dex. girc
// recovers the panic and the error is shown in the server buffer, where it
// can be noticed and reported.
func (c *Client) onHandlerPanic(_ *girc.Client, err *girc.HandlerError) {
	c.events.push(BufferNewMessageMsg{
		Server:    c.serverName,
		Buffer:    "",
		Timestamp: time.Now(),
		From:      "--",
		Text:      fmt.Sprintf("irc: internal error while handling %s: %v", err.Event.Command, err.Panic),
		Type:      MessageTypeServer,
	})
}
