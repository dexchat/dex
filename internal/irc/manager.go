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
func (m *ClientManager) Send(server, target, message string) error {
	return m.sendPrivmsg(server, target, message, false)
}

// SendAction sends a CTCP ACTION, the /me message, to a channel or nickname.
func (m *ClientManager) SendAction(server, target, action string) error {
	return m.sendPrivmsg(server, target, action, true)
}

// SendLines sends lines to a channel or nickname as separate messages, in
// order. Lines of another SendLines call to the same server are not mixed
// in; other messages may be. girc paces the lines, so SendLines blocks until
// the last one is sent. It returns how many lines were sent before an error.
func (m *ClientManager) SendLines(server, target string, lines []string) (int, error) {
	client, err := m.client(server)
	if err != nil {
		return 0, err
	}
	client.linesMu.Lock()
	defer client.linesMu.Unlock()
	for i, line := range lines {
		if err := m.sendPrivmsg(server, target, line, false); err != nil {
			return i, err
		}
	}
	return len(lines), nil
}

// sendPrivmsg sends text in as many PRIVMSGs as its length requires. A long
// action is sent as several actions.
func (m *ClientManager) sendPrivmsg(server, target, text string, action bool) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if girc.IsValidChannel(target) && !client.IsInChannel(target) {
		return fmt.Errorf("irc: not in channel %s", target)
	}

	echo := client.HasCapability("echo-message")
	for _, part := range splitMessage(text, client.maxMessageBytes(target, action)) {
		if action {
			part = encodeAction(part)
		}
		client.trackSent(target, part, echo)
		client.Cmd.Message(target, part)
	}
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

// Whois requests information about nick. The replies arrive as events for
// the server buffer.
func (m *ClientManager) Whois(server, nick string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if nick == "" || girc.IsValidChannel(nick) || strings.ContainsAny(nick, ", ") {
		return fmt.Errorf("irc: invalid nickname %s", nick)
	}

	client.Cmd.Whois(nick)
	return nil
}

// Nick requests a nickname change. The server confirms it with a NICK event
// or rejects it with an error numeric shown in the server buffer.
func (m *ClientManager) Nick(server, nick string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	if nick == "" || strings.Contains(nick, " ") || strings.HasPrefix(nick, ":") {
		return fmt.Errorf("irc: invalid nickname %s", nick)
	}

	client.Cmd.Nick(nick)
	return nil
}

// SetTopic changes channel's topic; an empty topic clears it. The server
// announces the change to the channel with a TOPIC event or rejects it with
// an error numeric.
func (m *ClientManager) SetTopic(server, channel, topic string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}

	client.Cmd.Topic(channel, topic)
	return nil
}

// Op gives channel operator status to each nick. The server announces each
// change to the channel with a MODE event or rejects it with an error numeric.
func (m *ClientManager) Op(server, channel string, nicks []string) error {
	client, err := m.connectedClient(server)
	if err != nil {
		return err
	}
	for _, nick := range nicks {
		if nick == "" || girc.IsValidChannel(nick) || strings.ContainsAny(nick, ", ") {
			return fmt.Errorf("irc: invalid nickname %s", nick)
		}
	}

	// One MODE per nick stays within the server's MODES limit.
	for _, nick := range nicks {
		client.Cmd.Mode(channel, "+o", nick)
	}
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
