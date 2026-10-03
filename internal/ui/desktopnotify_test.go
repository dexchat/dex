package ui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/notify"
)

type fakeNotifier struct {
	calls []notify.Notification
	err   error
}

func (f *fakeNotifier) Notify(_ context.Context, n notify.Notification) error {
	f.calls = append(f.calls, n)
	return f.err
}

// runCmd executes cmd and every command batched inside it, returning the
// resulting messages.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case nil:
		return nil
	case tea.BatchMsg:
		var msgs []tea.Msg
		for _, batched := range msg {
			msgs = append(msgs, runCmd(batched)...)
		}
		return msgs
	default:
		return []tea.Msg{msg}
	}
}

func countBells(msgs []tea.Msg) int {
	count := 0
	for _, msg := range msgs {
		if raw, ok := msg.(tea.RawMsg); ok && raw.Msg == "\a" {
			count++
		}
	}
	return count
}

func newDesktopTestModel(t *testing.T) (*Model, *fakeNotifier, *time.Time) {
	t.Helper()
	m := newActivityTestModel()
	m.config.Notifications.ShowBody = true
	fake := &fakeNotifier{}
	m.SetNotifier(fake)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time {
		return now
	}
	return m, fake, &now
}

func mentionIn(buffer string, at time.Time) irc.BufferNewMessageMsg {
	return irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    buffer,
		Timestamp: at,
		From:      "alice",
		Text:      "dexuser: ping",
		Type:      irc.MessageTypeNormal,
	}
}

func TestDesktopNotificationContent(t *testing.T) {
	channelMsg := irc.BufferNewMessageMsg{Buffer: "#go", From: "alice", Text: "hello"}
	tests := []struct {
		name     string
		msg      func(irc.BufferNewMessageMsg) irc.BufferNewMessageMsg
		showBody bool
		want     notify.Notification
	}{
		{
			name:     "channel message",
			showBody: true,
			want:     notify.Notification{Summary: "alice in #go", Body: "hello"},
		},
		{
			name: "direct message",
			msg: func(msg irc.BufferNewMessageMsg) irc.BufferNewMessageMsg {
				msg.DirectMessage = true
				msg.Buffer = "alice"
				return msg
			},
			showBody: true,
			want:     notify.Notification{Summary: "alice", Body: "hello"},
		},
		{
			name: "formatting removed",
			msg: func(msg irc.BufferNewMessageMsg) irc.BufferNewMessageMsg {
				msg.Text = "\x02hi\x02 \x0304,01there\x0f"
				return msg
			},
			showBody: true,
			want:     notify.Notification{Summary: "alice in #go", Body: "hi there"},
		},
		{
			name: "action in body",
			msg: func(msg irc.BufferNewMessageMsg) irc.BufferNewMessageMsg {
				msg.Action = true
				msg.Text = "waves"
				return msg
			},
			showBody: true,
			want:     notify.Notification{Summary: "alice in #go", Body: "* alice waves"},
		},
		{
			name: "hidden body hides action text",
			msg: func(msg irc.BufferNewMessageMsg) irc.BufferNewMessageMsg {
				msg.Action = true
				msg.Text = "shares a secret"
				return msg
			},
			showBody: false,
			want:     notify.Notification{Summary: "alice in #go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := channelMsg
			if tt.msg != nil {
				msg = tt.msg(msg)
			}
			if got := desktopNotification(msg, tt.showBody); got != tt.want {
				t.Fatalf("desktopNotification() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestProcessIncomingSendsDesktopNotificationIndependentOfSound(t *testing.T) {
	for _, sound := range []bool{true, false} {
		t.Run(fmt.Sprintf("sound=%v", sound), func(t *testing.T) {
			m, fake, now := newDesktopTestModel(t)
			m.config.Notifications.Sound = sound

			msgs := runCmd(m.processIncomingMessage(mentionIn("#random", *now)))

			want := []notify.Notification{{Summary: "alice in #random", Body: "dexuser: ping"}}
			if !slices.Equal(fake.calls, want) {
				t.Fatalf("Notify calls = %#v, want %#v", fake.calls, want)
			}
			wantBells := 0
			if sound {
				wantBells = 1
			}
			if got := countBells(msgs); got != wantBells {
				t.Fatalf("bells = %d, want %d", got, wantBells)
			}
		})
	}
}

func TestProcessIncomingSkipsDesktopNotification(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*Model, *irc.BufferNewMessageMsg)
	}{
		{
			name: "active buffer while focused",
			configure: func(m *Model, msg *irc.BufferNewMessageMsg) {
				m.activeBuffer = makeBufferKey("libera", msg.Buffer)
				m.terminalFocused = true
			},
		},
		{
			name: "old playback",
			configure: func(_ *Model, msg *irc.BufferNewMessageMsg) {
				msg.Timestamp = msg.Timestamp.Add(-maxNotificationAge - time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, fake, now := newDesktopTestModel(t)
			msg := mentionIn("#random", *now)
			tt.configure(m, &msg)

			runCmd(m.processIncomingMessage(msg))

			if len(fake.calls) != 0 {
				t.Fatalf("Notify calls = %#v, want none", fake.calls)
			}
		})
	}
}

func TestDesktopCooldownIsPerBufferWhileBellIsGlobal(t *testing.T) {
	m, fake, now := newDesktopTestModel(t)

	var msgs []tea.Msg
	msgs = append(msgs, runCmd(m.processIncomingMessage(mentionIn("#random", *now)))...)
	*now = now.Add(time.Second)
	msgs = append(msgs, runCmd(m.processIncomingMessage(mentionIn("#go", *now)))...)
	msgs = append(msgs, runCmd(m.processIncomingMessage(mentionIn("#random", *now)))...)

	var summaries []string
	for _, call := range fake.calls {
		summaries = append(summaries, call.Summary)
	}
	if got, want := strings.Join(summaries, ", "), "alice in #random, alice in #go"; got != want {
		t.Fatalf("notified buffers = %q, want %q", got, want)
	}
	if got := countBells(msgs); got != 1 {
		t.Fatalf("bells = %d, want 1", got)
	}

	*now = now.Add(m.config.Notifications.Cooldown)
	runCmd(m.processIncomingMessage(mentionIn("#random", *now)))
	if got := len(fake.calls); got != 3 {
		t.Fatalf("Notify calls after cooldown = %d, want 3", got)
	}
}

func TestDesktopUnavailableErrorDisablesNotifications(t *testing.T) {
	cfg := &config.Config{
		Notifications: config.Notifications{
			Events:   map[string]bool{config.NotificationMention: true},
			Cooldown: 2 * time.Second,
		},
		Servers: []*config.Server{
			{Name: "libera", Nickname: "dexuser", Channels: []string{"#random"}},
			{Name: "oftc", Nickname: "dexuser"},
		},
	}
	m := New(cfg)
	fake := &fakeNotifier{}
	m.SetNotifier(fake)
	libera := m.buffers[makeBufferKey("libera", "")]
	oftc := m.buffers[makeBufferKey("oftc", "")]
	libera.Chat.SetSize(80, 10)
	oftc.Chat.SetSize(80, 10)

	unavailable := fmt.Errorf("%w: connect to session bus: no bus", notify.ErrUnavailable)
	_, _ = m.Update(desktopNotificationResultMsg{server: "oftc", err: unavailable})
	_, _ = m.Update(desktopNotificationResultMsg{server: "oftc", err: unavailable})

	view := plainText(oftc.Chat.View())
	if got := strings.Count(view, "disabled until dex restarts"); got != 1 {
		t.Fatalf("oftc server buffer reports the error %d times, want once:\n%s", got, view)
	}
	if view := plainText(libera.Chat.View()); strings.Contains(view, "disabled") {
		t.Fatalf("error leaked into another server buffer:\n%s", view)
	}

	runCmd(m.processIncomingMessage(mentionIn("#random", time.Now())))
	if len(fake.calls) != 0 {
		t.Fatalf("Notify calls after disabling = %#v, want none", fake.calls)
	}
}

func TestDesktopCallErrorIsReportedOncePerFailureStreak(t *testing.T) {
	m, fake, now := newDesktopTestModel(t)
	server := m.buffers[makeBufferKey("libera", "")]
	server.Chat.SetSize(80, 10)

	failed := desktopNotificationResultMsg{server: "libera", err: errors.New("send desktop notification: daemon gone")}
	_, _ = m.Update(failed)
	_, _ = m.Update(failed)
	if got := strings.Count(plainText(server.Chat.View()), "daemon gone"); got != 1 {
		t.Fatalf("error reported %d times, want once", got)
	}

	runCmd(m.processIncomingMessage(mentionIn("#random", *now)))
	if got := len(fake.calls); got != 1 {
		t.Fatalf("Notify calls after a call error = %d, want 1", got)
	}

	_, _ = m.Update(desktopNotificationResultMsg{server: "libera"})
	_, _ = m.Update(failed)
	if got := strings.Count(plainText(server.Chat.View()), "daemon gone"); got != 2 {
		t.Fatalf("error after a success reported %d times in total, want 2", got)
	}
}
