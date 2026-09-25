package irc

import (
	"testing"

	"github.com/dexchat/dex/internal/config"
	"github.com/lrstanley/girc"
)

func TestDecodeAction(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantText   string
		wantAction bool
	}{
		{name: "action", text: "\x01ACTION waves\x01", wantText: "waves", wantAction: true},
		{name: "missing closing delimiter", text: "\x01ACTION waves", wantText: "waves", wantAction: true},
		{name: "empty action", text: "\x01ACTION\x01", wantText: "", wantAction: true},
		{name: "keeps inner spacing", text: "\x01ACTION  a  b \x01", wantText: " a  b ", wantAction: true},
		{name: "other CTCP command", text: "\x01ACTIONS\x01", wantText: "\x01ACTIONS\x01"},
		{name: "plain text", text: "ACTION waves", wantText: "ACTION waves"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, action := decodeAction(tt.text)
			if text != tt.wantText || action != tt.wantAction {
				t.Fatalf("decodeAction(%q) = %q, %v; want %q, %v", tt.text, text, action, tt.wantText, tt.wantAction)
			}
		})
	}
}

func TestIncomingActionIsDecoded(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.onEvent(client.Client, girc.Event{
		Command: girc.PRIVMSG,
		Source:  girc.ParseSource("alice!alice@example.test"),
		Params:  []string{"#go", "\x01ACTION waves\x01"},
	})

	messages := queuedMessages(client)
	if len(messages) != 1 {
		t.Fatalf("queued %d messages, want 1", len(messages))
	}
	if msg := messages[0]; !msg.Action || msg.Text != "waves" || msg.From != "alice" {
		t.Fatalf("incoming action = %+v, want alice's action \"waves\"", msg)
	}
}

func TestSentActionEchoIsOwnEcho(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})
	client.trackSent("#go", encodeAction("waves"), true)

	client.onEvent(client.Client, girc.Event{
		Command: girc.PRIVMSG,
		Source:  girc.ParseSource("tester!tester@example.test"),
		Params:  []string{"#go", "\x01ACTION waves\x01"},
		Echo:    true,
	})

	messages := queuedMessages(client)
	if len(messages) != 1 {
		t.Fatalf("queued %d messages, want 1", len(messages))
	}
	if msg := messages[0]; !msg.OwnEcho || !msg.Action || msg.Text != "waves" {
		t.Fatalf("action echo = %+v, want own echo of action \"waves\"", msg)
	}
}

func TestSentActionWithoutEchoMessageQueuesLocalAction(t *testing.T) {
	client := NewClient("libera", &config.Server{
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	})

	client.trackSent("#go", encodeAction("waves"), false)

	messages := queuedMessages(client)
	if len(messages) != 1 {
		t.Fatalf("queued %d local echoes, want 1", len(messages))
	}
	if msg := messages[0]; !msg.OwnEcho || !msg.Action || msg.Text != "waves" {
		t.Fatalf("local action echo = %+v, want own echo of action \"waves\"", msg)
	}
}
