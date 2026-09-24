package irc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lrstanley/girc"
	"github.com/dexchat/dex/internal/config"
)

func TestClientRegistersOnlyOneSynchronousHandler(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	handlers := reflect.ValueOf(client.Handlers).Elem().FieldByName("external")
	if !handlers.IsValid() {
		t.Fatal("girc Caller no longer exposes external handlers in the expected shape")
	}
	var commands []string
	for _, command := range handlers.MapKeys() {
		if handlers.MapIndex(command).Len() > 0 {
			commands = append(commands, command.String())
		}
	}
	if want := []string{girc.ALL_EVENTS}; !reflect.DeepEqual(commands, want) {
		t.Fatalf("handlers registered for %v, want only %v", commands, want)
	}

	ids := externalHandlerIDs(t, client, girc.ALL_EVENTS)
	if len(ids) != 1 || strings.HasSuffix(ids[0], ":bg") {
		t.Fatalf("ALL_EVENTS handlers = %v, want one synchronous handler", ids)
	}
}

func TestSortUserListUsesAdvertisedPrefixOrder(t *testing.T) {
	users := []string{
		"zed",
		"+voice",
		"!highest",
		"@operator",
		"alice",
		"~owner",
	}

	sortUserList(users, "(yqaohv)!~&@%+")

	want := []string{
		"!highest",
		"~owner",
		"@operator",
		"+voice",
		"alice",
		"zed",
	}
	if !reflect.DeepEqual(users, want) {
		t.Fatalf("unexpected user order: got %v, want %v", users, want)
	}
}

func TestMembershipEventsMarkUserListChanged(t *testing.T) {
	tests := []girc.Event{
		{Command: girc.JOIN, Source: girc.ParseSource("bob!bob@example.test"), Params: []string{"#go"}},
		{Command: girc.PART, Source: girc.ParseSource("alice!alice@example.test"), Params: []string{"#go"}},
		{Command: girc.KICK, Source: girc.ParseSource("op!op@example.test"), Params: []string{"#go", "alice"}},
		{Command: girc.QUIT, Source: girc.ParseSource("alice!alice@example.test"), Params: []string{"bye"}},
		{Command: girc.NICK, Source: girc.ParseSource("alice!alice@example.test"), Params: []string{"alice_"}},
		{Command: girc.MODE, Source: girc.ParseSource("op!op@example.test"), Params: []string{"#go", "+o", "alice"}},
		{Command: girc.RPL_ENDOFNAMES, Params: []string{"tester", "#go", "End of /NAMES list"}},
		{Command: girc.RPL_ENDOFWHO, Params: []string{"tester", "#go", "End of /WHO list"}},
	}

	for _, event := range tests {
		t.Run(event.Command, func(t *testing.T) {
			client := NewClient("testnet", &config.Server{
				Address:  "irc.example.test",
				Port:     6697,
				Nickname: "tester",
			})
			joinChannel(client, "alice", "#go")

			client.RunHandlers(&event)

			if got := changedUserLists(client); !reflect.DeepEqual(got, []string{"#go"}) {
				t.Fatalf("%s marked user lists %v, want [#go]", event.Command, got)
			}
		})
	}
}

func TestMembershipChangeIsQueuedOnlyAfterGircAppliesIt(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	// Without girc's own JOIN handling, the change has not been applied.
	client.onEvent(client.Client, girc.Event{
		Command: girc.JOIN,
		Source:  girc.ParseSource("bob!bob@example.test"),
		Params:  []string{"#go"},
	})
	if got := changedUserLists(client); len(got) != 0 {
		t.Fatalf("user list for %v queued before girc applied the JOIN", got)
	}

	client.onEvent(client.Client, girc.Event{Command: girc.UPDATE_STATE})
	if got := changedUserLists(client); !reflect.DeepEqual(got, []string{"#go"}) {
		t.Fatalf("user lists queued after UPDATE_STATE = %v, want [#go]", got)
	}
}

func TestUserModeDoesNotMarkUserListChanged(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onEvent(client.Client, girc.Event{
		Command: girc.MODE,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"tester", "+i"},
	})

	if got := changedUserLists(client); len(got) != 0 {
		t.Fatalf("user MODE marked user lists %v, want none", got)
	}
}

func TestSelfJoinIsQueuedForUI(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onJoin(client.Client, girc.Event{
		Command: girc.JOIN,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go"},
	})

	want := []Event{ChannelJoinedMsg{Server: "testnet", Channel: "#go"}}
	if got := queuedEvents(client); !reflect.DeepEqual(got, want) {
		t.Fatalf("queued self-JOIN = %#v, want %#v", got, want)
	}
}

func TestJoinErrorIsRoutedToServerBuffer(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onJoinError(client.Client, girc.Event{
		Command: girc.ERR_INVITEONLYCHAN,
		Params:  []string{"tester", "#private", "Cannot join channel (+i)"},
	})

	if got := len(queuedMessages(client)); got != 1 {
		t.Fatalf("join error queued %d messages, want 1", got)
	}
	msg := queuedMessages(client)[0]
	if msg.Buffer != "" {
		t.Fatalf("join error buffer = %q, want server buffer", msg.Buffer)
	}
	if !strings.Contains(msg.Text, "#private") {
		t.Fatalf("join error %q does not identify the rejected channel", msg.Text)
	}
}

func TestJoinErrorsAreQueuedForServerBuffer(t *testing.T) {
	for _, numeric := range []string{
		girc.ERR_NOSUCHCHANNEL,
		girc.ERR_TOOMANYCHANNELS,
		girc.ERR_BADCHANNELKEY,
		girc.ERR_BANNEDFROMCHAN,
		girc.ERR_CHANNELISFULL,
		girc.ERR_INVITEONLYCHAN,
		girc.ERR_BADCHANMASK,
	} {
		client := NewClient("testnet", &config.Server{
			Address:  "irc.example.test",
			Port:     6697,
			Nickname: "tester",
		})

		client.onEvent(client.Client, girc.Event{
			Command: numeric,
			Params:  []string{"tester", "#private", "Cannot join channel"},
		})

		if got := len(queuedMessages(client)); got != 1 {
			t.Errorf("JOIN error numeric %s queued %d messages, want 1", numeric, got)
		}
	}
}

func TestListReplyIsFormattedForServerBuffer(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onListReply(client.Client, girc.Event{
		Command: girc.RPL_LIST,
		Params:  []string{"tester", "#go", "42", "The Go channel"},
	})

	if got := len(queuedMessages(client)); got != 1 {
		t.Fatalf("LIST reply queued %d messages, want 1", got)
	}
	msg := queuedMessages(client)[0]
	if msg.Buffer != "" {
		t.Fatalf("LIST reply buffer = %q, want server buffer", msg.Buffer)
	}
	if want := "#go (42 users) The Go channel"; msg.Text != want {
		t.Fatalf("LIST reply = %q, want %q", msg.Text, want)
	}
}

func TestListRepliesAreQueuedForServerBuffer(t *testing.T) {
	for _, numeric := range []string{girc.RPL_LISTSTART, girc.RPL_LIST, girc.RPL_LISTEND, girc.ERR_TOOMANYMATCHES} {
		client := NewClient("testnet", &config.Server{
			Address:  "irc.example.test",
			Port:     6697,
			Nickname: "tester",
		})

		client.onEvent(client.Client, girc.Event{
			Command: numeric,
			Params:  []string{"tester", "#go", "42", "The Go channel"},
		})

		if got := len(queuedMessages(client)); got != 1 {
			t.Errorf("LIST numeric %s queued %d messages, want 1", numeric, got)
		}
	}
}

func TestSelfPartDoesNotCreateAChatMessage(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onPart(client.Client, girc.Event{
		Command: girc.PART,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go"},
	})

	want := []Event{ChannelPartedMsg{Server: "testnet", Channel: "#go"}}
	if got := queuedEvents(client); !reflect.DeepEqual(got, want) {
		t.Fatalf("self-PART queued %#v, want only %#v", got, want)
	}
}

func TestSelfKickDoesNotCreateAChatMessage(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onKick(client.Client, girc.Event{
		Command: girc.KICK,
		Source:  girc.ParseSource("op!op@example.test"),
		Params:  []string{"#go", "tester", "spamming"},
	})

	want := []Event{ChannelPartedMsg{Server: "testnet", Channel: "#go"}}
	if got := queuedEvents(client); !reflect.DeepEqual(got, want) {
		t.Fatalf("self-KICK queued %#v, want only %#v", got, want)
	}
}

func TestKickMessageIdentifiesKickerAndReason(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onKick(client.Client, girc.Event{
		Command: girc.KICK,
		Source:  girc.ParseSource("op!op@example.test"),
		Params:  []string{"#go", "Guest22", "spamming"},
	})

	if got := len(queuedMessages(client)); got != 1 {
		t.Fatalf("KICK queued %d messages, want 1", got)
	}
	msg := queuedMessages(client)[0]
	if msg.Buffer != "#go" {
		t.Fatalf("kick message buffer = %q, want #go", msg.Buffer)
	}
	if !msg.UserEvent {
		t.Fatal("kick message should be marked as a user event")
	}
	if want := "Guest22 has been kicked by op (spamming)"; msg.Text != want {
		t.Fatalf("kick message = %q, want %q", msg.Text, want)
	}
}

func TestKickMessageWithoutReasonOmitsParens(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onKick(client.Client, girc.Event{
		Command: girc.KICK,
		Source:  girc.ParseSource("op!op@example.test"),
		Params:  []string{"#go", "Guest22"},
	})

	if got := len(queuedMessages(client)); got != 1 {
		t.Fatalf("KICK queued %d messages, want 1", got)
	}
	if want := "Guest22 has been kicked by op"; queuedMessages(client)[0].Text != want {
		t.Fatalf("kick message = %q, want %q", queuedMessages(client)[0].Text, want)
	}
}

func TestJoinAndPartUserListChangesDoNotSendWhoRefresh(t *testing.T) {
	for _, command := range []string{girc.JOIN, girc.PART} {
		t.Run(command, func(t *testing.T) {
			var debug bytes.Buffer
			ircClient := girc.New(girc.Config{
				Server:     "irc.example.test",
				Nick:       "tester",
				User:       "tester",
				AllowFlood: true,
				Debug:      &debug,
			})
			client := &Client{
				Client:     ircClient,
				serverName: "testnet",
			}

			client.onUserListChange(ircClient, girc.Event{
				Command: command,
				Source:  girc.ParseSource("alice!alice@example.test"),
				Params:  []string{"#brasil"},
			})

			if strings.Contains(debug.String(), "WHO #brasil") {
				t.Fatalf("expected %s user-list change not to send WHO refresh, debug log:\n%s", command, debug.String())
			}
		})
	}
}

func TestNextSendsOneUserListPerChannelAfterOtherEvents(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	joinChannel(client, "alice", "#brasil")
	client.RunHandlers(&girc.Event{
		Command: girc.RPL_NAMREPLY,
		Params:  []string{"tester", "=", "#brasil", "@op alice"},
	})
	joinChannel(client, "bob", "#idlerpg")
	clearQueue(client)

	message := BufferNewMessageMsg{Server: "testnet", Buffer: "#brasil", Text: "hello"}
	client.queueUserListChanges("#brasil")
	client.events.push(message)
	client.queueUserListChanges("#Brasil", "#idlerpg")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	events, err := client.Next(ctx)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	want := []Event{
		message,
		UserListMsg{Server: "testnet", Channel: "#brasil", Users: []string{"@op", "alice"}, Prefixes: "@+"},
		UserListMsg{Server: "testnet", Channel: "#idlerpg", Users: []string{"bob"}, Prefixes: "@+"},
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("Next() = %#v, want %#v", events, want)
	}
}

func TestNextKeepsWaitingWhenOnlyUntrackedUserListsChanged(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	client.queueUserListChanges("#gone")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	events, err := client.Next(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Next() = %#v, %v; want to keep waiting until the deadline", events, err)
	}
}

func TestNickMarksOnlyTheUsersChannels(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	joinChannel(client, "alice", "#go")
	joinChannel(client, "bob", "#rust")

	client.RunHandlers(&girc.Event{
		Command: girc.NICK,
		Source:  girc.ParseSource("alice!alice@example.test"),
		Params:  []string{"alice_"},
	})

	if got := changedUserLists(client); !reflect.DeepEqual(got, []string{"#go"}) {
		t.Fatalf("NICK marked user lists %v, want [#go]", got)
	}
}

func TestClientDisablesGircAutoJoinQueries(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	if !client.Config.DisableAutoWHOOnJoin {
		t.Fatal("expected dex to disable girc automatic WHO on self-JOIN")
	}
	if !client.Config.DisableAutoMODEOnJoin {
		t.Fatal("expected dex to disable girc automatic MODE on self-JOIN")
	}
}

func TestPendingMessageKeyNormalizesServerTargetAndMessage(t *testing.T) {
	got := pendingMessageKey("Libera", "#Brasil", "  hello from dex  ")
	want := pendingMessageKey("libera", "#brasil", "hello from dex")

	if got != want {
		t.Fatalf("pending message keys should match: got %q, want %q", got, want)
	}
}

func TestOwnMessagesFromTheServerArriveAsEchoes(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "johnbogle",
	})
	client.trackSent("#brasil", "hello from dex", true)
	client.trackSent("alice", "hello alice", true)

	// girc marks messages from our nickname as echoes in its read loop, so
	// feed them through a real connection instead of RunHandlers.
	server := mockServer(t, client)
	for _, line := range []string{
		":johnbogle!jdex@example.test PRIVMSG #Brasil :hello from dex",
		":johnbogle!jdex@example.test PRIVMSG alice :hello alice",
		":johnbogle!jdex@example.test PRIVMSG #brasil :sent from another client",
	} {
		if _, err := io.WriteString(server, line+"\r\n"); err != nil {
			t.Fatalf("write %q: %v", line, err)
		}
	}
	waitFor(t, "three messages", func() bool { return len(queuedMessages(client)) == 3 })

	messages := queuedMessages(client)
	if msg := messages[0]; msg.Buffer != "#Brasil" || !msg.OwnEcho {
		t.Fatalf("channel echo = %+v, want OwnEcho in #Brasil", msg)
	}
	if msg := messages[1]; msg.Buffer != "alice" || !msg.DirectMessage || !msg.OwnEcho {
		t.Fatalf("direct echo = %+v, want OwnEcho in the recipient's buffer", msg)
	}
	if msg := messages[2]; msg.OwnEcho {
		t.Fatalf("message from another client = %+v, want it shown", msg)
	}
}

func TestRepeatedSendsAreEachRecognized(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	client.trackSent("#go", "ok", true)
	client.trackSent("#go", "ok", true)

	for range 3 {
		client.onEvent(client.Client, girc.Event{
			Command: girc.PRIVMSG,
			Source:  girc.ParseSource("tester!tester@example.test"),
			Params:  []string{"#go", "ok"},
			Echo:    true,
		})
	}

	var ownEchoes []bool
	for _, msg := range queuedMessages(client) {
		ownEchoes = append(ownEchoes, msg.OwnEcho)
	}
	if want := []bool{true, true, false}; !reflect.DeepEqual(ownEchoes, want) {
		t.Fatalf("OwnEcho for three echoes of two sends = %v, want %v", ownEchoes, want)
	}
}

func TestSendWithoutEchoMessageQueuesLocalEcho(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.trackSent("#go", "hi", false)
	client.trackSent("alice", "oi", false)

	messages := queuedMessages(client)
	if len(messages) != 2 {
		t.Fatalf("queued %d local echoes, want 2", len(messages))
	}
	if msg := messages[0]; msg.Buffer != "#go" || msg.DirectMessage || !msg.OwnEcho || msg.From != "tester" || msg.Text != "hi" {
		t.Fatalf("channel local echo = %+v", msg)
	}
	if msg := messages[1]; msg.Buffer != "alice" || !msg.DirectMessage || !msg.OwnEcho {
		t.Fatalf("direct local echo = %+v", msg)
	}
	if client.sent.consume(pendingMessageKey("libera", "#go", "hi"), time.Now()) {
		t.Fatal("a send without echo-message should not wait for an echo")
	}
}

func TestQuitIsRoutedToChannelsGircIsAboutToForget(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	quitter := girc.ParseSource("Guest22!~Guest22@2804:1e68:c211:45f3:5485:907e:2a08:1c7b")

	// Use girc's real dispatch so its state tracking runs alongside onEvent.
	client.RunHandlers(&girc.Event{Command: girc.JOIN, Source: quitter, Params: []string{"#brasil"}})
	client.RunHandlers(&girc.Event{Command: girc.QUIT, Source: quitter, Params: []string{"Quit: Client closed"}})

	if user := client.LookupUser("Guest22"); user != nil {
		t.Fatalf("girc still tracks the quitting user in %v", user.ChannelList)
	}

	var quits []BufferNewMessageMsg
	for _, msg := range queuedMessages(client) {
		if strings.Contains(msg.Text, "has quit") {
			quits = append(quits, msg)
		}
	}
	if len(quits) != 1 {
		t.Fatalf("queued %d quit messages, want 1: %#v", len(quits), quits)
	}
	if quits[0].Buffer != "#brasil" {
		t.Fatalf("quit message buffer = %q, want #brasil", quits[0].Buffer)
	}
	if !quits[0].UserEvent {
		t.Fatal("quit message should be marked as a user event")
	}
}

func TestNickUpdateUsesNicknameFromEvent(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	// Call onEvent directly: girc's own RPL_WELCOME handler would rename us and
	// dispatch CONNECTED from another goroutine.
	client.onEvent(client.Client, girc.Event{
		Command: girc.RPL_WELCOME,
		Source:  girc.ParseSource("irc.example.test"),
		Params:  []string{"tester_", "Welcome to the network"},
	})
	client.RunHandlers(&girc.Event{
		Command: girc.NICK,
		Source:  girc.ParseSource("alice!alice@example.test"),
		Params:  []string{"alice_"},
	})
	client.RunHandlers(&girc.Event{
		Command: girc.NICK,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"dexter"},
	})

	var nicks []string
	for _, event := range queuedEvents(client) {
		if update, ok := event.(NickUpdateMsg); ok {
			nicks = append(nicks, update.Nick)
		}
	}
	if want := []string{"tester_", "dexter"}; !reflect.DeepEqual(nicks, want) {
		t.Fatalf("nickname updates = %v, want %v", nicks, want)
	}
}

func TestEchoMessageIsMarkedAsOwnEcho(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	client.trackSent("#go", "hello", true)

	client.onEvent(client.Client, girc.Event{
		Command: girc.PRIVMSG,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go", "hello"},
		Echo:    true,
	})

	messages := queuedMessages(client)
	if len(messages) != 1 {
		t.Fatalf("echo queued %d messages, want 1", len(messages))
	}
	if !messages[0].OwnEcho || messages[0].Buffer != "#go" {
		t.Fatalf("echo message = %+v, want OwnEcho in #go", messages[0])
	}
}

func TestSelfPartIsQueuedAfterEarlierChannelMessages(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onPrivmsg(client.Client, girc.Event{
		Command: girc.PRIVMSG,
		Source:  girc.ParseSource("alice!alice@example.test"),
		Params:  []string{"#go", "last words"},
	})
	client.onPart(client.Client, girc.Event{
		Command: girc.PART,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go"},
	})

	events := queuedEvents(client)
	if len(events) != 2 {
		t.Fatalf("queued %d events, want 2: %#v", len(events), events)
	}
	if msg, ok := events[0].(BufferNewMessageMsg); !ok || msg.Text != "last words" {
		t.Fatalf("first event = %#v, want the channel message", events[0])
	}
	if _, ok := events[1].(ChannelPartedMsg); !ok {
		t.Fatalf("second event = %#v, want the self-PART", events[1])
	}
}

// joinChannel adds nick to channel in girc's state through its real JOIN
// handling, then clears the events that JOIN queued.
func joinChannel(client *Client, nick, channel string) {
	client.RunHandlers(&girc.Event{
		Command: girc.JOIN,
		Source:  girc.ParseSource(nick + "!" + nick + "@example.test"),
		Params:  []string{channel},
	})
	clearQueue(client)
}

func clearQueue(client *Client) {
	client.events.mu.Lock()
	defer client.events.mu.Unlock()
	client.events.events = nil
}

// changedUserLists returns the channels of queued userListChanged markers.
func changedUserLists(client *Client) []string {
	var channels []string
	for _, event := range queuedEvents(client) {
		if changed, ok := event.(userListChanged); ok {
			channels = append(channels, changed.channel)
		}
	}
	return channels
}

// mockServer connects client to an in-memory server and returns the server
// side. Everything the client sends is discarded. The connection is closed
// when the test ends.
func mockServer(t *testing.T, client *Client) net.Conn {
	t.Helper()
	server, conn := net.Pipe()
	done := make(chan struct{})
	go func() {
		_ = client.MockConnect(conn)
		close(done)
	}()
	go func() { _, _ = io.Copy(io.Discard, server) }()
	t.Cleanup(func() {
		server.Close()
		<-done
	})
	return server
}

// queuedEvents returns the client's queued events without waiting for the
// batching window.
func queuedEvents(client *Client) []Event {
	client.events.mu.Lock()
	defer client.events.mu.Unlock()
	return append([]Event(nil), client.events.events...)
}

func queuedMessages(client *Client) []BufferNewMessageMsg {
	var messages []BufferNewMessageMsg
	for _, event := range queuedEvents(client) {
		if msg, ok := event.(BufferNewMessageMsg); ok {
			messages = append(messages, msg)
		}
	}
	return messages
}

func externalHandlerIDs(t *testing.T, client *Client, command string) []string {
	t.Helper()

	handlers := reflect.ValueOf(client.Handlers).Elem().FieldByName("external")
	if !handlers.IsValid() {
		t.Fatal("girc Caller no longer exposes external handlers in the expected shape")
	}

	commandHandlers := handlers.MapIndex(reflect.ValueOf(command))
	if !commandHandlers.IsValid() {
		return nil
	}

	ids := make([]string, 0, commandHandlers.Len())
	for _, key := range commandHandlers.MapKeys() {
		ids = append(ids, key.String())
	}
	return ids
}

func TestTLSConfigVerifiesCertificatesByDefault(t *testing.T) {
	conf := tlsConfig(&config.Server{Address: "irc.example.test", Port: 6697})
	if conf.InsecureSkipVerify {
		t.Fatal("expected certificate verification to be enabled by default")
	}
	if conf.ServerName != "irc.example.test" {
		t.Fatalf("ServerName = %q, want irc.example.test", conf.ServerName)
	}
}

func TestTLSConfigSkipsVerificationWhenConfigured(t *testing.T) {
	conf := tlsConfig(&config.Server{Address: "znc.example.test", Port: 6697, SSLSkipVerify: true})
	if !conf.InsecureSkipVerify {
		t.Fatal("expected certificate verification to be skipped")
	}
}
