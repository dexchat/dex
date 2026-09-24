package irc

import (
	"context"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dexchat/dex/internal/config"
	"github.com/lrstanley/girc"
)

// newLocalClient returns a plain-text client for a server on localhost.
func newLocalClient(t *testing.T, port int) *Client {
	t.Helper()
	ssl := false
	return NewClient("libera", &config.Server{
		Address:  "127.0.0.1",
		Port:     port,
		SSL:      &ssl,
		Nickname: "tester",
		Channels: []string{"#go"},
	})
}

// closedPort returns a localhost port with nothing listening on it.
func closedPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

func waitFor(t *testing.T, what string, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !done() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestBackoffResetsAfterRegisteredConnection(t *testing.T) {
	retry := backoff{delays: []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}}

	var got []time.Duration
	for _, registered := range []bool{false, false, false, false, true, false} {
		got = append(got, retry.next(registered))
	}

	want := []time.Duration{
		10 * time.Second, 20 * time.Second, 40 * time.Second, 40 * time.Second,
		10 * time.Second, 20 * time.Second,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reconnect delays = %v, want %v", got, want)
	}
}

func TestConnectLoopReportsRetryAndStopsWhenCanceled(t *testing.T) {
	client := newLocalClient(t, closedPort(t))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		client.connectLoop(ctx, []time.Duration{time.Hour})
		close(done)
	}()

	waitFor(t, "the retry message", func() bool { return len(queuedMessages(client)) > 0 })
	msg := queuedMessages(client)[0]
	if msg.Buffer != "" || msg.Type != MessageTypeDisconnected {
		t.Fatalf("retry message = %+v, want a disconnected message in the server buffer", msg)
	}
	if !strings.HasSuffix(msg.Text, "reconnecting in 3600 seconds...") {
		t.Fatalf("retry message = %q, want the reconnect delay", msg.Text)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect loop kept waiting after its context was canceled")
	}
}

func TestConnectLoopStopsAfterClose(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = io.Copy(io.Discard, conn)
	}()

	client := newLocalClient(t, listener.Addr().(*net.TCPAddr).Port)
	// girc sets up Close only after the socket connects, without a lock.
	// INITIALIZED is dispatched after that setup, so closing then is safe.
	initialized := make(chan struct{})
	var once sync.Once
	client.Handlers.Add(girc.INITIALIZED, func(*girc.Client, girc.Event) {
		once.Do(func() { close(initialized) })
	})
	done := make(chan struct{})
	go func() {
		client.connectLoop(context.Background(), []time.Duration{time.Hour})
		close(done)
	}()

	select {
	case <-initialized:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the connection")
	}
	client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("connect loop did not stop after Close")
	}
	for _, msg := range queuedMessages(client) {
		if strings.Contains(msg.Text, "reconnecting") {
			t.Fatalf("connect loop scheduled a reconnect after Close: %q", msg.Text)
		}
	}
}

func TestChannelCaseIsUpdatedOnlyForOwnJoin(t *testing.T) {
	client := newLocalClient(t, closedPort(t))

	client.onEvent(client.Client, girc.Event{
		Command: girc.JOIN,
		Source:  girc.ParseSource("alice!alice@example.test"),
		Params:  []string{"#Go"},
	})
	if got := channelNameUpdates(client); len(got) != 0 {
		t.Fatalf("another user's JOIN sent channel name updates %v", got)
	}

	client.onEvent(client.Client, girc.Event{
		Command: girc.JOIN,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#Go"},
	})
	if got := channelNameUpdates(client); !reflect.DeepEqual(got, []string{"#Go"}) {
		t.Fatalf("own JOIN sent channel name updates %v, want [#Go]", got)
	}
	if client.channels[0] != "#go" {
		t.Fatalf("configured channels changed to %v", client.channels)
	}
}

func channelNameUpdates(client *Client) []string {
	var names []string
	for _, event := range queuedEvents(client) {
		if update, ok := event.(ChannelNameUpdateMsg); ok {
			names = append(names, update.CanonicalName)
		}
	}
	return names
}

func TestManagerCommandsReportErrorsInsteadOfQueueingThem(t *testing.T) {
	manager := NewClientManager([]*config.Server{{
		Name:     "libera",
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}})
	defer manager.DisconnectAll("")

	commands := map[string]func(server string) error{
		"Send": func(server string) error { return manager.Send(server, "#go", "hello") },
		"Join": func(server string) error { return manager.Join(server, "#go", "") },
		"Part": func(server string) error { return manager.Part(server, "#go", "") },
		"List": func(server string) error { return manager.List(server, "") },
	}
	for name, run := range commands {
		if err := run("libera"); err == nil || !strings.Contains(err.Error(), "not connected to libera") {
			t.Errorf("%s while disconnected error = %v, want not connected", name, err)
		}
		if err := run("missing"); err == nil || !strings.Contains(err.Error(), "unknown server missing") {
			t.Errorf("%s for an unknown server error = %v, want unknown server", name, err)
		}
	}

	if events := queuedEvents(manager.clients["libera"]); len(events) != 0 {
		t.Fatalf("command errors were queued as events: %#v", events)
	}
}
