package irc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lrstanley/girc"
	"github.com/dexchat/dex/internal/config"
)

// ClientManager handles multiple clients/server connections
// With more than one connection, the UI wouldn't know which channel to send a message
type ClientManager struct {
	clients map[string]*Client

	// ctx is canceled by DisconnectAll to release pending Next calls.
	ctx    context.Context
	cancel context.CancelFunc

	// loops tracks the connection loops started by ConnectAll, so
	// DisconnectAll can wait for QUIT to reach the servers.
	loops sync.WaitGroup
}

// quitTimeout limits how long DisconnectAll waits for connections to end
// after sending QUIT.
const quitTimeout = 2 * time.Second

func NewClientManager(servers []*config.Server) *ClientManager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &ClientManager{
		clients: make(map[string]*Client),
		ctx:     ctx,
		cancel:  cancel,
	}
	for _, server := range servers {
		m.clients[server.Name] = NewClient(server.Name, server)
	}
	return m
}

// Next blocks until the named server has events for the UI. It returns an
// error for an unknown server or after DisconnectAll.
func (m *ClientManager) Next(server string) ([]Event, error) {
	client, err := m.client(server)
	if err != nil {
		return nil, err
	}
	return client.Next(m.ctx)
}

// ConnectAll starts one connection loop per server. The loops reconnect after
// failures and stop when DisconnectAll is called.
func (m *ClientManager) ConnectAll() {
	for _, client := range m.clients {
		m.loops.Go(func() {
			client.connectLoop(m.ctx, reconnectDelays)
		})
	}
}

// Send sends a message to a channel or nickname. The returned error describes
// why it could not be sent; errors from the server arrive as events instead.
func (m *ClientManager) Send(server, channel, message string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if girc.IsValidChannel(channel) && !client.IsInChannel(channel) {
		return fmt.Errorf("irc: not in channel %s", channel)
	}

	client.trackSent(channel, message, client.HasCapability("echo-message"))
	client.Cmd.Message(channel, message)
	return nil
}

func (m *ClientManager) Part(server, channel, reason string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if !girc.IsValidChannel(channel) {
		return errors.New("irc: /leave is only available in a channel")
	}
	if !client.IsInChannel(channel) {
		return fmt.Errorf("irc: not in channel %s", channel)
	}

	if reason == "" {
		client.Cmd.Part(channel)
		return nil
	}
	client.Cmd.PartMessage(channel, reason)
	return nil
}

func (m *ClientManager) Join(server, channel, key string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if !girc.IsValidChannel(channel) || strings.Contains(channel, ",") {
		return fmt.Errorf("irc: invalid channel %s", channel)
	}
	if client.IsInChannel(channel) {
		return fmt.Errorf("irc: already in channel %s", channel)
	}

	if key == "" {
		client.Cmd.Join(channel)
		return nil
	}
	client.Cmd.JoinKey(channel, key)
	return nil
}

func (m *ClientManager) List(server, channel string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if channel != "" && (!girc.IsValidChannel(channel) || strings.Contains(channel, ",")) {
		return fmt.Errorf("irc: invalid channel %s", channel)
	}

	if channel == "" {
		client.Cmd.List()
		return nil
	}
	client.Cmd.List(channel)
	return nil
}

func (m *ClientManager) client(server string) (*Client, error) {
	client, ok := m.clients[server]
	if !ok {
		return nil, fmt.Errorf("irc: unknown server %s", server)
	}
	return client, nil
}

func (m *ClientManager) connectedClient(server string) (*Client, error) {
	client, err := m.client(server)
	if err != nil {
		return nil, err
	}
	if !client.IsConnected() {
		return nil, fmt.Errorf("irc: not connected to %s", server)
	}
	return client, nil
}

// DisconnectAll stops every connection loop. Connected clients send QUIT
// with reason, and girc closes each connection once its QUIT is written.
// Connections that have not ended after quitTimeout are closed without it.
func (m *ClientManager) DisconnectAll(reason string) {
	m.cancel()
	for _, client := range m.clients {
		if !client.IsConnected() {
			client.Close()
			continue
		}
		// Quit can block while girc's send queue is full, so it must not
		// hold up the timeout below.
		go client.Quit(reason)
	}

	done := make(chan struct{})
	go func() {
		m.loops.Wait()
		close(done)
	}()
	timer := time.NewTimer(quitTimeout)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	}

	for _, client := range m.clients {
		client.Close()
	}
}
