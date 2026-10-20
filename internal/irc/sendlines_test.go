package irc

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dexchat/dex/internal/config"
	"github.com/lrstanley/girc"
)

// TestConcurrentSendLinesAreNotInterleaved sends two pastes at once. Each
// must reach the server as one block, in order.
func TestConcurrentSendLinesAreNotInterleaved(t *testing.T) {
	manager := NewClientManager([]*config.Server{{
		Name:     "libera",
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}})
	client := manager.clients["libera"]
	// This test checks ordering, not pacing.
	client.Config.AllowFlood = true
	lines := sentLines(t, client)
	waitFor(t, "the connection", client.IsConnected)

	const perPaste = 100
	pastes := map[string][]string{}
	for _, name := range []string{"a", "b"} {
		for i := range perPaste {
			pastes[name] = append(pastes[name], fmt.Sprintf("%s%03d", name, i))
		}
	}
	var wg sync.WaitGroup
	for name, paste := range pastes {
		wg.Go(func() {
			if sent, err := manager.SendLines("libera", "alice", paste); err != nil || sent != perPaste {
				t.Errorf("SendLines(%s) = %d, %v; want %d, nil", name, sent, err, perPaste)
			}
		})
	}

	var got []string
	for len(got) < 2*perPaste {
		select {
		case line := <-lines:
			if event := girc.ParseEvent(line); event != nil && event.Command == girc.PRIVMSG {
				got = append(got, event.Last())
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("received %d of %d lines", len(got), 2*perPaste)
		}
	}
	wg.Wait()

	first := pastes[got[0][:1]]
	second := pastes[got[perPaste][:1]]
	want := strings.Join(append(append([]string(nil), first...), second...), " ")
	if strings.Join(got, " ") != want {
		t.Fatalf("pastes were interleaved or reordered:\n%s", strings.Join(got, " "))
	}
}

func TestSendLinesReportsLinesSentBeforeAnError(t *testing.T) {
	manager := NewClientManager([]*config.Server{{
		Name:     "libera",
		Address:  "irc.example.test",
		Port:     6697,
		Nickname: "tester",
	}})

	sent, err := manager.SendLines("libera", "alice", []string{"one", "two"})
	if sent != 0 || err == nil || !strings.Contains(err.Error(), "not connected to libera") {
		t.Fatalf("SendLines while disconnected = %d, %v; want 0 and not connected", sent, err)
	}
}
