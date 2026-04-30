package irc

import (
	"reflect"
	"strings"
	"testing"

	"github.com/lrstanley/girc"
	"github.com/vaaleyard/dex/internal/config"
)

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
