package irc

import (
	"bufio"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dexchat/dex/internal/config"
	"github.com/lrstanley/girc"
)

func TestSplitMessage(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		limit int
		want  []string
	}{
		{name: "fits", text: "hello world", limit: 11, want: []string{"hello world"}},
		{name: "breaks at last space", text: "aaa bbb ccc", limit: 8, want: []string{"aaa bbb", "ccc"}},
		{name: "space right after limit", text: "aaaa bbbb", limit: 4, want: []string{"aaaa", "bbbb"}},
		{name: "no space", text: "abcdefgh", limit: 3, want: []string{"abc", "def", "gh"}},
		{name: "keeps runes whole", text: "ééé", limit: 3, want: []string{"é", "é", "é"}},
		{name: "rune larger than limit", text: "🙂a", limit: 2, want: []string{"🙂", "a"}},
		{name: "drops trailing space part", text: "aaaa ", limit: 4, want: []string{"aaaa"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitMessage(tt.text, tt.limit); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitMessage(%q, %d) = %q, want %q", tt.text, tt.limit, got, tt.want)
			}
		})
	}
}

// TestLongMessagesAreSentInPartsGircKeepsWhole checks that every part dex
// sends is short enough that girc does not split it again. Each part is
// tracked on its own, so its echo is recognized as this client's.
func TestLongMessagesAreSentInPartsGircKeepsWhole(t *testing.T) {
	for _, action := range []bool{false, true} {
		t.Run(map[bool]string{false: "message", true: "action"}[action], func(t *testing.T) {
			manager := NewClientManager([]*config.Server{{
				Name:     "libera",
				Address:  "irc.example.test",
				Port:     6697,
				Nickname: "tester",
			}})
			client := manager.clients["libera"]
			// This test checks splitting, not pacing; girc's flood
			// limiter would make it wait seconds between parts.
			client.Config.AllowFlood = true
			lines := sentLines(t, client)
			waitFor(t, "the connection", client.IsConnected)

			text := strings.TrimSpace(strings.Repeat("palavra ação 🙂 ", 80))
			send := manager.Send
			if action {
				send = manager.SendAction
			}
			if err := send("libera", "alice", text); err != nil {
				t.Fatalf("send: %v", err)
			}

			// A part longer than girc's limit would arrive as several PRIVMSGs
			// for one tracked part, so the echo count below would not match.
			var parts []string
			for strings.Join(parts, " ") != text {
				var line string
				select {
				case line = <-lines:
				case <-time.After(2 * time.Second):
					t.Fatalf("timed out after receiving parts %q", parts)
				}
				event := girc.ParseEvent(line)
				if event == nil || event.Command != girc.PRIVMSG {
					continue
				}
				if event.LenOpts(false) >= client.MaxEventLength() {
					t.Fatalf("part of %d bytes would be split by girc", event.LenOpts(false))
				}
				part, isAction := decodeAction(event.Last())
				if isAction != action {
					t.Fatalf("part %q action = %v, want %v", event.Last(), isAction, action)
				}
				parts = append(parts, part)
			}
			if len(parts) < 2 {
				t.Fatalf("sent %d part, want the message split", len(parts))
			}

			echoes := queuedMessages(client)
			if len(echoes) != len(parts) {
				t.Fatalf("queued %d local echoes for %d parts", len(echoes), len(parts))
			}
			for i, echo := range echoes {
				if !echo.OwnEcho || echo.Action != action || echo.Text != parts[i] {
					t.Fatalf("echo %d = %+v, want own echo of %q", i, echo, parts[i])
				}
			}
		})
	}
}

// sentLines connects client to an in-memory server and returns the lines the
// client writes to it.
func sentLines(t *testing.T, client *Client) <-chan string {
	t.Helper()
	server, conn := net.Pipe()
	done := make(chan struct{})
	go func() {
		_ = client.MockConnect(conn)
		close(done)
	}()
	lines := make(chan string, 64)
	go func() {
		scanner := bufio.NewScanner(server)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()
	t.Cleanup(func() {
		server.Close()
		<-done
	})
	return lines
}
