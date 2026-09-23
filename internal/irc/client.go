package irc

import (
	"context"
	"crypto/tls"
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

// Next blocks until this server has events for the UI and returns every
// event queued so far, in order. Bursts are collected into one batch.
func (c *Client) Next(ctx context.Context) ([]Event, error) {
	return c.events.next(ctx, eventBatchWindow)
}

func (c *Client) addHandlers() {
	c.Handlers.Add(girc.ALL_EVENTS, c.onEvent)
}
