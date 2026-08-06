package ui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/history"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui/components/users"
)

func TestPaneForMouseWheelUsesPaneBounds(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want scrollPane
	}{
		{name: "channels", x: 0, want: scrollPaneChannels},
		{name: "last channel column", x: channelsPanelMaxWidth - 1, want: scrollPaneChannels},
		{name: "chat", x: channelsPanelMaxWidth, want: scrollPaneChat},
		{name: "users", x: 100 - usersPanelMaxWidth, want: scrollPaneUsers},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paneForMouseWheel(tea.MouseWheelMsg{
				X:      tt.x,
				Button: tea.MouseWheelDown,
			}, 100)
			if got != tt.want {
				t.Fatalf("paneForMouseWheel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaneForMouseWheelIgnoresNonWheelMouse(t *testing.T) {
	got := paneForMouseWheel(tea.MouseClickMsg{
		X:      0,
		Button: tea.MouseLeft,
	}, 100)
	if got != scrollPaneNone {
		t.Fatalf("paneForMouseWheel() = %v, want %v", got, scrollPaneNone)
	}
}

func TestPlaybackBatchIsSplitIntoResponsiveChunks(t *testing.T) {
	messages := make(irc.BufferNewMessageBatchMsg, maxPlaybackMessagesPerUpdate+1)
	current, remaining := playbackChunk(messages)
	if got := len(current); got != maxPlaybackMessagesPerUpdate {
		t.Fatalf("current playback chunk has %d messages, want %d", got, maxPlaybackMessagesPerUpdate)
	}
	if got := len(remaining); got != 1 {
		t.Fatalf("remaining playback chunk has %d messages, want 1", got)
	}
}

func TestInactiveBufferMessageIncrementsUnreadActivity(t *testing.T) {
	m := newActivityTestModel()

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello there",
	})

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got, want := buf.UnreadCount, 1; got != want {
		t.Fatalf("UnreadCount = %d, want %d", got, want)
	}
	if got, want := buf.MentionCount, 0; got != want {
		t.Fatalf("MentionCount = %d, want %d", got, want)
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); !strings.Contains(content, "1") {
		t.Fatalf("expected unread badge in channel list, got:\n%s", content)
	}
}

func TestInactiveBufferMentionIncrementsMentionActivity(t *testing.T) {
	m := newActivityTestModel()

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hey dexuser, can you check this?",
	})

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got, want := buf.UnreadCount, 1; got != want {
		t.Fatalf("UnreadCount = %d, want %d", got, want)
	}
	if got, want := buf.MentionCount, 1; got != want {
		t.Fatalf("MentionCount = %d, want %d", got, want)
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); !strings.Contains(content, "@1") {
		t.Fatalf("expected mention badge in channel list, got:\n%s", content)
	}
}

func TestActiveBufferMessagesAndOwnEchoesDoNotIncrementActivity(t *testing.T) {
	m := newActivityTestModel()
	m.activeBuffer = makeBufferKey("libera", "#random")

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hey dexuser",
	})
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "dexuser",
		Text:      "my own message",
		OwnEcho:   true,
	})

	if got := m.buffers[makeBufferKey("libera", "#random")].UnreadCount; got != 0 {
		t.Fatalf("active buffer UnreadCount = %d, want 0", got)
	}
	if got := m.buffers[makeBufferKey("libera", "#go")].UnreadCount; got != 0 {
		t.Fatalf("own echo UnreadCount = %d, want 0", got)
	}
}

func TestInactiveUserListRendersWhenBufferIsSelected(t *testing.T) {
	m := newActivityTestModel()
	_, _ = m.Update(irc.UserListMsg{
		Server:  "libera",
		Channel: "#random",
		Users:   []string{"alice"},
	})

	random := m.buffers[makeBufferKey("libera", "#random")]
	if random.Users.HasUser("alice") {
		t.Fatal("inactive user list should not render during playback")
	}

	m.activeBuffer = random.Key
	random.Users, _ = random.Users.Update(users.UserListMsg(random.members))
	if !random.Users.HasUser("alice") {
		t.Fatal("selected buffer should render its saved user list")
	}
}

func TestNonNormalMessagesDoNotIncrementActivity(t *testing.T) {
	m := newActivityTestModel()

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "--",
		Text:      "irc: connected",
		Type:      irc.MessageTypeConnected,
	})

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got := buf.UnreadCount; got != 0 {
		t.Fatalf("UnreadCount = %d, want 0", got)
	}
	if got := buf.MentionCount; got != 0 {
		t.Fatalf("MentionCount = %d, want 0", got)
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); strings.Contains(content, "1") {
		t.Fatalf("expected no badge for non-normal message, got:\n%s", content)
	}
}

func TestNewBuildsChannelListWithoutLoadingChannelHistory(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	writeTestHistory(t, "libera", "#go", "stored message")

	cfg := &config.Config{
		Servers: []*config.Server{
			{
				Name:     "libera",
				Nickname: "dexuser",
				Channels: []string{"#go"},
			},
		},
	}

	m := New(cfg)
	buf := m.buffers[makeBufferKey("libera", "#go")]
	if buf == nil {
		t.Fatal("expected configured channel buffer to exist")
	}
	if got := len(buf.History.Entries()); got != 0 {
		t.Fatalf("New loaded %d history entries synchronously, want 0", got)
	}

	m.channels = m.channels.SetSize(channelsPanelMaxWidth, 20)
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); !strings.Contains(content, "#go") {
		t.Fatalf("expected configured channel to render before history load, got:\n%s", content)
	}
}

func TestConfiguredHistoryLoadPopulatesExistingBuffers(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	writeTestHistory(t, "libera", "#go", "stored message")

	cfg := &config.Config{
		Servers: []*config.Server{
			{
				Name:     "libera",
				Nickname: "dexuser",
				Channels: []string{"#go"},
			},
		},
	}

	m := New(cfg)
	msg := m.loadConfiguredHistory()().(initialHistoryLoadedMsg)
	m.Update(msg)

	buf := m.buffers[makeBufferKey("libera", "#go")]
	if got := len(buf.History.Entries()); got != 1 {
		t.Fatalf("loaded history entries = %d, want 1", got)
	}
}

func TestDiscoveredChannelHistoryLoadsOnlyWhenRequested(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeTestHistory(t, "libera", "#go", "stored message")

	m := New(&config.Config{Servers: []*config.Server{{Name: "libera", Nickname: "dexuser"}}})
	buf, createCmd := m.getOrCreateBuffer("libera", "#go")
	if got := len(buf.History.Entries()); got != 0 {
		t.Fatalf("new buffer loaded %d history entries synchronously, want 0", got)
	}
	if createCmd == nil {
		t.Fatal("expected a sidebar creation command")
	}
	if buf.historyLoading {
		t.Fatal("discovered channel should not start disk history loading at startup")
	}

	m.Update(m.requestHistoryLoad(buf)())
	if got := len(buf.History.Entries()); got != 1 {
		t.Fatalf("loaded history entries = %d, want 1", got)
	}
}

func writeTestHistory(t *testing.T, server, buffer, text string) {
	t.Helper()

	log := history.NewLog()
	log.Insert(history.LogEntry{
		ReceivedAt: time.Now().UnixNano(),
		ServerTime: time.Now().UnixNano(),
		Username:   "alice",
		Text:       text,
	})
	if err := log.Flush(server, buffer); err != nil {
		t.Fatalf("failed to write test history: %v", err)
	}
}

func TestSelectingBufferClearsActivity(t *testing.T) {
	m := newActivityTestModel()

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "dexuser: ping",
	})

	if got := m.buffers[makeBufferKey("libera", "#go")].MentionCount; got != 1 {
		t.Fatalf("MentionCount before selecting = %d, want 1", got)
	}

	_, _ = m.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})

	buf := m.buffers[makeBufferKey("libera", "#go")]
	if got := buf.UnreadCount; got != 0 {
		t.Fatalf("UnreadCount after selecting = %d, want 0", got)
	}
	if got := buf.MentionCount; got != 0 {
		t.Fatalf("MentionCount after selecting = %d, want 0", got)
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); strings.Contains(content, "@1") {
		t.Fatalf("expected mention badge to clear, got:\n%s", content)
	}
}

func TestUnreadBadgeCanBeDisabledGlobally(t *testing.T) {
	m := newActivityTestModel()
	m.config.UI.UnreadBadges = false

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello there",
	})

	if got := m.buffers[makeBufferKey("libera", "#random")].UnreadCount; got != 1 {
		t.Fatalf("UnreadCount = %d, want 1", got)
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); strings.Contains(content, "1") {
		t.Fatalf("expected unread badge to be hidden, got:\n%s", content)
	}
}

func TestMentionBadgeCanBeDisabledPerServerWhileUnreadRemains(t *testing.T) {
	m := newActivityTestModel()
	m.config.Servers[0].MentionBadges = boolPtr(false)

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hey dexuser, can you check this?",
	})

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got, want := buf.UnreadCount, 1; got != want {
		t.Fatalf("UnreadCount = %d, want %d", got, want)
	}
	if got, want := buf.MentionCount, 1; got != want {
		t.Fatalf("MentionCount = %d, want %d", got, want)
	}
	content := plainText(m.channels.View(channelsPanelMaxWidth, 20))
	if strings.Contains(content, "@1") {
		t.Fatalf("expected mention badge to be hidden, got:\n%s", content)
	}
	if !strings.Contains(content, "1") {
		t.Fatalf("expected unread badge to remain visible, got:\n%s", content)
	}
}

func TestMentionDetectionRequiresWholeNickToken(t *testing.T) {
	if !messageMentionsNick("hey dexuser: ping", "dexuser") {
		t.Fatal("expected exact nick token to count as mention")
	}
	if messageMentionsNick("hey superdexuser ping", "dexuser") {
		t.Fatal("expected substring inside a longer token not to count as mention")
	}
	if !messageMentionsNick("DEXUSER, ping", "dexuser") {
		t.Fatal("expected mention matching to be case-insensitive")
	}
}

func TestShouldSoundNotification(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*Model, *Buffer, *irc.BufferNewMessageMsg)
		want      bool
	}{
		{
			name: "inactive mention",
			want: true,
		},
		{
			name: "inactive direct message",
			configure: func(_ *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				msg.DirectMessage = true
				msg.Text = "hello"
			},
			want: true,
		},
		{
			name: "ordinary channel message",
			configure: func(_ *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				msg.Text = "hello"
			},
			want: false,
		},
		{
			name: "active buffer mention",
			configure: func(m *Model, buf *Buffer, _ *irc.BufferNewMessageMsg) {
				m.activeBuffer = buf.Key
			},
			want: false,
		},
		{
			name: "own echo",
			configure: func(_ *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				msg.OwnEcho = true
			},
			want: false,
		},
		{
			name: "server message",
			configure: func(_ *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				msg.Type = irc.MessageTypeServer
			},
			want: false,
		},
		{
			name: "sound disabled",
			configure: func(m *Model, _ *Buffer, _ *irc.BufferNewMessageMsg) {
				m.config.Notifications.Sound = false
			},
			want: false,
		},
		{
			name: "mention event disabled",
			configure: func(m *Model, _ *Buffer, _ *irc.BufferNewMessageMsg) {
				m.config.Notifications.Events[config.NotificationMention] = false
			},
			want: false,
		},
		{
			name: "direct message event disabled",
			configure: func(m *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				m.config.Notifications.Events[config.NotificationDirectMessage] = false
				msg.DirectMessage = true
				msg.Text = "hello"
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newActivityTestModel()
			buf := m.buffers[makeBufferKey("libera", "#random")]
			msg := irc.BufferNewMessageMsg{
				Server: "libera",
				Buffer: "#random",
				From:   "alice",
				Text:   "dexuser: ping",
				Type:   irc.MessageTypeNormal,
			}
			if tt.configure != nil {
				tt.configure(m, buf, &msg)
			}

			if got := m.shouldSoundNotification(buf, msg); got != tt.want {
				t.Fatalf("shouldSoundNotification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newActivityTestModel() *Model {
	cfg := &config.Config{
		UI: config.UI{
			UnreadBadges:  true,
			MentionBadges: true,
		},
		Notifications: config.Notifications{
			Sound: true,
			Events: map[string]bool{
				config.NotificationMention:       true,
				config.NotificationDirectMessage: true,
			},
			Cooldown: 2 * time.Second,
		},
		Servers: []*config.Server{
			{
				Name:     "libera",
				Nickname: "dexuser",
				Channels: []string{"#go", "#random"},
			},
		},
	}
	m := New(cfg)
	m.width = 100
	m.height = 20
	m.channels = m.channels.SetSize(channelsPanelMaxWidth, 20)
	return m
}

func boolPtr(v bool) *bool {
	return &v
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func plainText(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}
