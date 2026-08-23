package irc

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
)

func TestMessageHandlersAreSynchronous(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	for _, command := range []string{girc.PRIVMSG, girc.NOTICE, girc.ALL_EVENTS} {
		t.Run(command, func(t *testing.T) {
			handlers := externalHandlerIDs(t, client, command)
			if len(handlers) == 0 {
				t.Fatalf("expected %s handler to be registered", command)
			}

			for _, id := range handlers {
				if strings.HasSuffix(id, ":bg") {
					t.Fatalf("%s handler %q is registered as background; order-sensitive message handlers must be synchronous", command, id)
				}
			}
		})
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

func TestMembershipEventsTriggerUserListRefresh(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	for _, command := range []string{girc.JOIN, girc.PART, girc.NICK, girc.MODE, girc.RPL_ENDOFNAMES, girc.RPL_ENDOFWHO} {
		t.Run(command, func(t *testing.T) {
			for _, id := range externalHandlerIDs(t, client, command) {
				if strings.HasSuffix(id, ":bg") {
					return
				}
			}
			t.Fatalf("expected %s to have a background user-list refresh handler, got %v", command, externalHandlerIDs(t, client, command))
		})
	}
}

func TestSelfJoinIsQueuedForAsynchronousUIDelivery(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	client.onJoin(client.Client, girc.Event{
		Command: girc.JOIN,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go"},
	})

	client.channelQueueMu.Lock()
	defer client.channelQueueMu.Unlock()
	if got := client.channelQueue; len(got) != 1 || got[0] != (ChannelJoinedMsg{Server: "testnet", Channel: "#go"}) {
		t.Fatalf("queued self-JOIN = %#v, want testnet/#go", got)
	}
}

func TestJoinErrorIsRoutedToServerBuffer(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)
	client.flushPending = true

	client.onJoinError(client.Client, girc.Event{
		Command: girc.ERR_INVITEONLYCHAN,
		Params:  []string{"tester", "#private", "Cannot join channel (+i)"},
	})

	if got := len(client.messageQueue); got != 1 {
		t.Fatalf("join error queued %d messages, want 1", got)
	}
	msg := client.messageQueue[0]
	if msg.Buffer != "" {
		t.Fatalf("join error buffer = %q, want server buffer", msg.Buffer)
	}
	if !strings.Contains(msg.Text, "#private") {
		t.Fatalf("join error %q does not identify the rejected channel", msg.Text)
	}
}

func TestJoinErrorHandlersCoverProtocolFailures(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	for _, numeric := range []string{
		girc.ERR_NOSUCHCHANNEL,
		girc.ERR_TOOMANYCHANNELS,
		girc.ERR_BADCHANNELKEY,
		girc.ERR_BANNEDFROMCHAN,
		girc.ERR_CHANNELISFULL,
		girc.ERR_INVITEONLYCHAN,
		girc.ERR_BADCHANMASK,
	} {
		if got := externalHandlerIDs(t, client, numeric); len(got) == 0 {
			t.Errorf("expected JOIN error handler for numeric %s", numeric)
		}
	}
}

func TestListReplyIsFormattedForServerBuffer(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)
	client.flushPending = true

	client.onListReply(client.Client, girc.Event{
		Command: girc.RPL_LIST,
		Params:  []string{"tester", "#go", "42", "The Go channel"},
	})

	if got := len(client.messageQueue); got != 1 {
		t.Fatalf("LIST reply queued %d messages, want 1", got)
	}
	msg := client.messageQueue[0]
	if msg.Buffer != "" {
		t.Fatalf("LIST reply buffer = %q, want server buffer", msg.Buffer)
	}
	if want := "#go (42 users) The Go channel"; msg.Text != want {
		t.Fatalf("LIST reply = %q, want %q", msg.Text, want)
	}
}

func TestListReplyHandlersAreRegistered(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	for _, numeric := range []string{girc.RPL_LISTSTART, girc.RPL_LIST, girc.RPL_LISTEND, girc.ERR_TOOMANYMATCHES} {
		if got := externalHandlerIDs(t, client, numeric); len(got) == 0 {
			t.Errorf("expected LIST reply handler for numeric %s", numeric)
		}
	}
}

func TestSelfPartDoesNotCreateAChatMessage(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	client.onPart(client.Client, girc.Event{
		Command: girc.PART,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go"},
	})

	if got := len(client.messageQueue); got != 0 {
		t.Fatalf("self-PART created %d chat messages, want 0", got)
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
				Client:       ircClient,
				serverName:   "testnet",
				userChannels: make(map[string]map[string]struct{}),
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

func TestUserListRefreshRequestsAreCoalescedByChannel(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	client.scheduleUserListRefresh(client.Client, "#brasil")
	client.scheduleUserListRefresh(client.Client, "#Brasil")
	client.scheduleUserListRefresh(client.Client, "#idlerpg")

	client.userListRefreshMu.Lock()
	defer client.userListRefreshMu.Unlock()

	if !client.userListRefreshScheduled {
		t.Fatal("expected user-list refresh to be scheduled")
	}
	if got := len(client.userListRefreshPending); got != 2 {
		t.Fatalf("expected refreshes to be coalesced per channel, got %d pending channels", got)
	}
}

func TestNickUserListChangeSchedulesRefreshDespiteNonChannelParam(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)
	client.onUserListChange(client.Client, girc.Event{
		Command: girc.NICK,
		Source:  girc.ParseSource("alice!alice@example.test"),
		Params:  []string{"alice_"},
	})

	client.userListRefreshMu.Lock()
	defer client.userListRefreshMu.Unlock()

	if !client.userListRefreshScheduled {
		t.Fatal("expected nick change to schedule a user-list refresh")
	}
}

func TestClientDisablesGircAutoJoinQueries(t *testing.T) {
	client := NewClient("testnet", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	if !client.Config.DisableAutoWhoOnJoin {
		t.Fatal("expected dex to disable girc automatic WHO on self-JOIN")
	}
	if !client.Config.DisableAutoModeOnJoin {
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

func TestPrivmsgFromSelfConsumesPendingMessage(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "johnbogle",
	}, nil)
	client.flushPending = true
	client.pendingMessages.Store(pendingMessageKey("libera", "#brasil", "hello from dex"), struct{}{})

	client.onPrivmsg(client.Client, girc.Event{
		Command: girc.PRIVMSG,
		Source:  girc.ParseSource("johnbogle!jdex@example.test"),
		Params:  []string{"#Brasil", "hello from dex"},
	})

	if _, pending := client.pendingMessages.Load(pendingMessageKey("libera", "#brasil", "hello from dex")); pending {
		t.Fatal("expected self PRIVMSG to consume the pending message")
	}

	if len(client.messageQueue) != 1 {
		t.Fatalf("expected one queued message, got %d", len(client.messageQueue))
	}
	if !client.messageQueue[0].OwnEcho {
		t.Fatal("expected self PRIVMSG matching a pending send to be marked OwnEcho")
	}
}

func TestDirectPrivmsgFromSelfUsesRecipientBuffer(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "johnbogle",
	}, nil)
	client.flushPending = true
	client.pendingMessages.Store(pendingMessageKey("libera", "alice", "hello alice"), struct{}{})

	client.onPrivmsg(client.Client, girc.Event{
		Command: girc.PRIVMSG,
		Source:  girc.ParseSource("johnbogle!jdex@example.test"),
		Params:  []string{"alice", "hello alice"},
	})

	if len(client.messageQueue) != 1 {
		t.Fatalf("expected one queued message, got %d", len(client.messageQueue))
	}
	msg := client.messageQueue[0]
	if msg.Buffer != "alice" {
		t.Fatalf("direct self-echo buffer = %q, want alice", msg.Buffer)
	}
	if !msg.DirectMessage || !msg.OwnEcho {
		t.Fatalf("direct self-echo = %+v, want DirectMessage and OwnEcho", msg)
	}
}

func TestUserChannelSnapshotTracksAndForgetsMembership(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)

	client.setChannelUsers("#brasil", []string{"Guest22", "johnbogle"})
	client.setChannelUsers("#idlerpg", []string{"guest22"})

	channels := client.channelsForUser("GUEST22")
	want := []string{"#brasil", "#idlerpg"}
	if !reflect.DeepEqual(channels, want) {
		t.Fatalf("channelsForUser() = %v, want %v", channels, want)
	}

	client.forgetUser("Guest22")
	if channels := client.channelsForUser("Guest22"); len(channels) != 0 {
		t.Fatalf("expected Guest22 to be removed from snapshot, still in %v", channels)
	}
}

func TestQuitUsesMembershipSnapshotWhenGircStateIsAlreadyDeleted(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}, nil)
	client.flushPending = true
	client.setChannelUsers("#brasil", []string{"Guest22"})

	client.onQuit(client.Client, girc.Event{
		Command: girc.QUIT,
		Source:  girc.ParseSource("Guest22!~Guest22@2804:1e68:c211:45f3:5485:907e:2a08:1c7b"),
		Params:  []string{"Quit: Client closed"},
	})

	if len(client.messageQueue) != 1 {
		t.Fatalf("expected one quit message, got %d", len(client.messageQueue))
	}
	if client.messageQueue[0].Buffer != "#brasil" {
		t.Fatalf("quit message buffer = %q, want #brasil", client.messageQueue[0].Buffer)
	}
	if !client.messageQueue[0].UserEvent {
		t.Fatal("quit message should be marked as a user event")
	}
	if channels := client.channelsForUser("Guest22"); len(channels) != 0 {
		t.Fatalf("expected quit user to be removed from snapshot, still in %v", channels)
	}
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
