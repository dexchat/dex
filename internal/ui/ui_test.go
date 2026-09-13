package ui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/commands"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/history"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/components/chat"
	"github.com/dexchat/dex/internal/ui/components/palette"
	"github.com/dexchat/dex/internal/ui/components/users"
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

func TestPaletteChannelSelectionUpdatesActiveBufferAndSidebarCursor(t *testing.T) {
	m := newActivityTestModel()

	_, _ = m.Update(palette.ChannelSelectionMsg{Server: "libera", Channel: "#go"})

	if got, want := m.activeBuffer, makeBufferKey("libera", "#go"); got != want {
		t.Fatalf("activeBuffer = %q, want %q", got, want)
	}
	selected := m.channels.Selected()
	if selected.Server != "libera" || selected.Channel != "#go" {
		t.Fatalf("sidebar selection = %#v, want libera/#go", selected)
	}
}

func TestGoToChannelKeybindingOpensChannelPickerDirectly(t *testing.T) {
	m := newActivityTestModel()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	_, _ = m.Update(tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl})

	if !m.palette.IsVisible() {
		t.Fatal("ctrl+g should open the channel picker")
	}
	view := plainText(m.palette.View())
	if !strings.Contains(view, "Search channels...") || !strings.Contains(view, "#go") {
		t.Fatalf("ctrl+g opened the wrong palette view:\n%s", view)
	}
}

func TestLastBufferKeybindingTogglesBetweenRecentBuffers(t *testing.T) {
	m := newActivityTestModel()
	var cmds []tea.Cmd
	m.selectBuffer("libera", "#go", &cmds)

	lastBufferKey := tea.KeyPressMsg{Code: '6', Mod: tea.ModCtrl}
	_, _ = m.Update(lastBufferKey)
	if got, want := m.activeBuffer, makeBufferKey("libera", ""); got != want {
		t.Fatalf("first %q activeBuffer = %q, want %q", lastBufferKey.String(), got, want)
	}

	_, _ = m.Update(tea.KeyPressMsg{Code: '^', Mod: tea.ModCtrl})
	if got, want := m.activeBuffer, makeBufferKey("libera", "#go"); got != want {
		t.Fatalf("second last-buffer shortcut activeBuffer = %q, want %q", got, want)
	}
	selected := m.channels.Selected()
	if selected.Server != "libera" || selected.Channel != "#go" {
		t.Fatalf("sidebar selection = %#v, want libera/#go", selected)
	}
}

func TestSelfPartRemovesActiveChannelAndSelectsServer(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = m.Update(irc.ChannelPartedMsg{Server: "libera", Channel: "#go"})

	if _, exists := m.buffers[channelKey]; exists {
		t.Fatal("self-PART should remove the channel buffer")
	}
	if got, want := m.activeBuffer, makeBufferKey("libera", ""); got != want {
		t.Fatalf("activeBuffer = %q, want %q", got, want)
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); strings.Contains(content, "#go") {
		t.Fatalf("self-PART should remove the channel from the sidebar:\n%s", content)
	}
}

func TestSelfPartPersistsDirtyChannelBeforeRemovingBuffer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	key := makeBufferKey("libera", "#go")
	m.activeBuffer = key
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello",
	})

	_, cmd := m.Update(irc.ChannelPartedMsg{Server: "libera", Channel: "#go"})
	if cmd == nil {
		t.Fatal("dirty self-PART should schedule persistence")
	}
	if _, exists := m.buffers[key]; !exists {
		t.Fatal("channel should remain until persistence completes")
	}

	_, _ = m.Update(cmd())
	if _, exists := m.buffers[key]; exists {
		t.Fatal("channel should be removed after persistence completes")
	}
}

func TestLeaveDoesNotRemoveChannelBeforeServerConfirmation(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = m.Update(chat.SendMessageMsg{Text: "/leave"})

	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("/leave should keep the channel until the server confirms PART")
	}
}

func TestPartIsAliasForLeave(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = m.Update(chat.SendMessageMsg{Text: "/part"})

	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("/part should keep the channel until server confirmation")
	}
}

func TestCloseRemovesPrivateBufferAndSelectsServer(t *testing.T) {
	m := newActivityTestModel()
	_, _ = m.getOrCreateBuffer("libera", "alice")
	m.directMessages.Add("libera", "alice")
	m.activeBuffer = makeBufferKey("libera", "alice")

	_, _ = m.Update(chat.SendMessageMsg{Text: "/close"})

	if _, exists := m.buffers[makeBufferKey("libera", "alice")]; exists {
		t.Fatal("/close should remove the private buffer")
	}
	if got, want := m.activeBuffer, makeBufferKey("libera", ""); got != want {
		t.Fatalf("activeBuffer = %q, want %q", got, want)
	}
	for _, directMessage := range m.directMessages.Users {
		if directMessage.Server == "libera" && directMessage.User == "alice" {
			t.Fatalf("closed direct message is still persisted: %#v", m.directMessages.Users)
		}
	}
}

func TestClosePersistsDirtyPrivateBufferBeforeRemovingIt(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	key := makeBufferKey("libera", "alice")
	_, _ = m.getOrCreateBuffer("libera", "alice")
	m.activeBuffer = key
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "alice",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello",
	})

	_, cmd := m.Update(chat.SendMessageMsg{Text: "/close"})
	if cmd == nil {
		t.Fatal("dirty /close should schedule persistence")
	}
	if _, exists := m.buffers[key]; !exists {
		t.Fatal("private buffer should remain until persistence completes")
	}

	_, _ = m.Update(cmd())
	if _, exists := m.buffers[key]; exists {
		t.Fatal("private buffer should be removed after persistence completes")
	}
}

func TestCloseDoesNotCloseChannel(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = m.Update(chat.SendMessageMsg{Text: "/close"})

	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("/close should not remove a channel buffer")
	}
}

func TestJoinDoesNotCreateChannelBeforeServerConfirmation(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#new")

	_, _ = m.Update(chat.SendMessageMsg{Text: "/join #new"})

	if _, exists := m.buffers[channelKey]; exists {
		t.Fatal("/join should not create a channel before the server confirms JOIN")
	}

	_, _ = m.Update(irc.ChannelJoinedMsg{Server: "libera", Channel: "#new"})
	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("self-JOIN confirmation should create the channel buffer")
	}
}

func TestJoinWithoutChannelShowsUsageInsteadOfCreatingBuffer(t *testing.T) {
	m := newActivityTestModel()
	active := m.getActiveBuffer()
	active.Chat.SetSize(80, 10)

	_, _ = m.Update(chat.SendMessageMsg{Text: "/join"})

	if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /join <channel> [key]") {
		t.Fatalf("missing /join usage error:\n%s", view)
	}
}

func TestListWithTooManyArgumentsShowsUsage(t *testing.T) {
	m := newActivityTestModel()
	active := m.getActiveBuffer()
	active.Chat.SetSize(80, 10)

	_, _ = m.Update(chat.SendMessageMsg{Text: "/list #go #rust"})

	if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /list [channel]") {
		t.Fatalf("missing /list usage error:\n%s", view)
	}
}

func TestMsgCreatesAndSelectsPrivateBuffer(t *testing.T) {
	m := newActivityTestModel()
	serverBuffer := m.getActiveBuffer()

	_, _ = m.Update(chat.SendMessageMsg{Text: "/msg Alice hello there"})

	privateKey := makeBufferKey(serverBuffer.Server, "Alice")
	privateBuffer, exists := m.buffers[privateKey]
	if !exists {
		t.Fatal("/msg should create a private buffer")
	}
	if got, want := m.activeBuffer, privateKey; got != want {
		t.Fatalf("activeBuffer = %q, want %q", got, want)
	}
	if view := plainText(privateBuffer.Chat.View()); !strings.Contains(view, "hello there") {
		t.Fatalf("private buffer should contain the outgoing message:\n%s", view)
	}
}

func TestMsgWithoutTargetOrMessageShowsUsage(t *testing.T) {
	m := newActivityTestModel()
	active := m.getActiveBuffer()
	active.Chat.SetSize(80, 10)

	_, _ = m.Update(chat.SendMessageMsg{Text: "/msg Alice"})

	if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /msg <user> <message>") {
		t.Fatalf("missing /msg usage error:\n%s", view)
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

func TestUserEventsCanSkipUnreadActivity(t *testing.T) {
	m := newActivityTestModel()
	m.config.UI.UnreadOnUserEvents = false

	for _, text := range []string{"alice has joined", "alice has left", "alice has quit"} {
		m.processIncomingMessage(irc.BufferNewMessageMsg{
			Server:    "libera",
			Buffer:    "#random",
			Timestamp: time.Now(),
			From:      "<--",
			Text:      text,
			UserEvent: true,
		})
	}

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got := buf.UnreadCount; got != 0 {
		t.Fatalf("UnreadCount = %d, want 0", got)
	}
	if got := len(buf.History.Entries()); got != 3 {
		t.Fatalf("history messages = %d, want 3", got)
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

func TestConfiguredChannelMessageShowsPersistentAndTemporaryNotificationState(t *testing.T) {
	m := newActivityTestModel()
	m.config.Servers[0].NotifyChannels = []string{"#random"}

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "deployment finished",
	})

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got, want := buf.NotificationCount, 1; got != want {
		t.Fatalf("NotificationCount = %d, want %d", got, want)
	}
	view := plainText(m.channels.View(channelsPanelMaxWidth, 20))
	if !strings.Contains(view, "libera/#random · alice") {
		t.Fatalf("expected notification source in sidebar footer, got:\n%s", view)
	}

	version := m.notificationNoticeVersion
	_, _ = m.Update(notificationNoticeExpiredMsg{version: version - 1})
	if view := plainText(m.channels.View(channelsPanelMaxWidth, 20)); !strings.Contains(view, "libera/#random · alice") {
		t.Fatal("an older timer cleared the current notification notice")
	}
	_, _ = m.Update(notificationNoticeExpiredMsg{version: version})
	if view := plainText(m.channels.View(channelsPanelMaxWidth, 20)); strings.Contains(view, "libera/#random · alice") {
		t.Fatal("notification notice did not clear after its current timer expired")
	}

	m.clearBufferActivity(buf.Key)
	if got := buf.NotificationCount; got != 0 {
		t.Fatalf("NotificationCount after clearing = %d, want 0", got)
	}
}

func TestAttentionPulseIncludesMentionsOutsideNotificationChannels(t *testing.T) {
	m := newActivityTestModel()
	buf := m.buffers[makeBufferKey("libera", "#random")]

	cmd := m.startAttentionPulse(buf, irc.BufferNewMessageMsg{
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "dexuser: can you check this?",
	})
	if cmd == nil {
		t.Fatal("expected a mention to start the attention pulse")
	}

	cmd = m.startAttentionPulse(buf, irc.BufferNewMessageMsg{
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "ordinary message",
	})
	if cmd != nil {
		t.Fatal("ordinary messages outside notify_channels should not pulse")
	}

	buf.MentionCount = 1
	cmd = m.startAttentionPulse(buf, irc.BufferNewMessageMsg{
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "bob",
		Text:      "follow-up without another mention",
	})
	if cmd == nil {
		t.Fatal("expected new activity to pulse while the buffer has an unread mention")
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
	random.Users, _ = random.Users.Update(users.UserListMsg{Users: random.members})
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

func TestConfiguredHistoryLoadOnlyLoadsReadState(t *testing.T) {
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
	if got := len(buf.History.Entries()); got != 0 {
		t.Fatalf("loaded history entries = %d, want 0", got)
	}
	if m.readState == nil {
		t.Fatal("read state was not loaded before startup connection")
	}
}

func TestPersistedDirectMessagesRestoreAfterServerConnects(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	directMessages := &history.DirectMessages{}
	directMessages.Add("libera", "alice")
	directMessages.Add("oftc", "bob")
	if err := directMessages.Flush(); err != nil {
		t.Fatalf("failed to persist direct messages: %v", err)
	}

	m := New(&config.Config{Servers: []*config.Server{
		{Name: "libera", Nickname: "dexuser"},
		{Name: "oftc", Nickname: "dexuser"},
	}})

	if _, ok := m.buffers[makeBufferKey("libera", "alice")]; ok {
		t.Fatal("persisted direct message restored before the server connected")
	}

	_, cmd := m.Update(irc.BufferNewMessageBatchMsg{{
		Server: "libera",
		Type:   irc.MessageTypeConnected,
	}})
	if cmd != nil {
		cmd()
	}

	if _, ok := m.buffers[makeBufferKey("libera", "alice")]; !ok {
		t.Fatal("persisted direct message was not restored after the server connected")
	}
	if _, ok := m.buffers[makeBufferKey("oftc", "bob")]; ok {
		t.Fatal("connecting one server restored another server's direct message")
	}

	_, cmd = m.Update(irc.BufferNewMessageBatchMsg{{
		Server: "libera",
		Type:   irc.MessageTypeConnected,
	}})
	if cmd != nil {
		cmd()
	}
	if got := len(m.buffers); got != 3 {
		t.Fatalf("buffer count after reconnect = %d, want 3", got)
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

func TestReadPlaybackDoesNotIncrementActivity(t *testing.T) {
	m := newActivityTestModel()
	readAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	m.readState.MarkRead("libera", "#random", history.ReadMarker{ServerTime: readAt.UnixNano()})

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: readAt,
		From:      "alice",
		Text:      "already read",
	})
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: readAt.Add(time.Second),
		From:      "alice",
		Text:      "new message",
	})

	buf := m.buffers[makeBufferKey("libera", "#random")]
	if got, want := buf.UnreadCount, 1; got != want {
		t.Fatalf("UnreadCount = %d, want %d", got, want)
	}
}

func TestReadMarkerSurvivesRestartPlayback(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	messageAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	first := newActivityTestModel()
	first.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: messageAt,
		From:      "alice",
		Text:      "read before restart",
	})
	key := makeBufferKey("libera", "#random")
	first.activeBuffer = key
	first.clearBufferActivity(key)
	first.flushAllHistory()

	second := newActivityTestModel()
	_, _ = second.Update(second.loadConfiguredHistory()())
	// Discovered buffers load disk history lazily, so playback can arrive before
	// duplicate detection has their prior entries available.
	second.buffers[key].History = history.NewLog()
	second.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: messageAt,
		From:      "alice",
		Text:      "read before restart",
	})
	second.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: messageAt.Add(time.Second),
		From:      "alice",
		Text:      "new after restart",
	})

	if got, want := second.buffers[key].UnreadCount, 1; got != want {
		t.Fatalf("UnreadCount after restart playback = %d, want %d", got, want)
	}
}

func TestSelectingBufferPersistsReadMarker(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	messageAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: messageAt,
		From:      "alice",
		Text:      "hello",
	})

	key := makeBufferKey("libera", "#random")
	m.activeBuffer = key
	m.clearBufferActivity(key)
	m.flushAllHistory()

	state, err := history.LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState() error = %v", err)
	}
	marker, ok := state.Marker("libera", "#random")
	if !ok || marker.ServerTime != messageAt.UnixNano() {
		t.Fatalf("Marker() = (%+v, %v), want timestamp %d", marker, ok, messageAt.UnixNano())
	}
}

func TestShutdownPersistsActiveBufferReadMarker(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	key := makeBufferKey("libera", "#random")
	m.activeBuffer = key
	messageAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: messageAt,
		From:      "alice",
		Text:      "hello",
	})
	m.flushAllHistory()

	state, err := history.LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState() error = %v", err)
	}
	marker, ok := state.Marker("libera", "#random")
	if !ok || marker.ServerTime != messageAt.UnixNano() {
		t.Fatalf("Marker() = (%+v, %v), want timestamp %d", marker, ok, messageAt.UnixNano())
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
	if !messageMentionsNick("de\x02xuser: formatted ping", "dexuser") {
		t.Fatal("expected IRC formatting inside nickname not to hide mention")
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
			name: "active buffer mention when focused",
			configure: func(m *Model, buf *Buffer, _ *irc.BufferNewMessageMsg) {
				m.activeBuffer = buf.Key
				m.terminalFocused = true
			},
			want: false,
		},
		{
			name: "active buffer mention when blurred",
			configure: func(m *Model, buf *Buffer, _ *irc.BufferNewMessageMsg) {
				m.activeBuffer = buf.Key
				m.terminalFocused = false
			},
			want: true,
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
		{
			name: "ignored direct message sender",
			configure: func(m *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				m.config.Servers[0].IgnoreDirectMessagesFrom = []string{"AlertBot"}
				msg.DirectMessage = true
				msg.From = "alertbot"
				msg.Text = "hello dexuser"
			},
			want: false,
		},
		{
			name: "configured notification channel",
			configure: func(m *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				m.config.Servers[0].NotifyChannels = []string{"#Alerts"}
				msg.Buffer = "#alerts"
				msg.Text = "ordinary message"
			},
			want: true,
		},
		{
			name: "unconfigured notification channel",
			configure: func(m *Model, _ *Buffer, msg *irc.BufferNewMessageMsg) {
				m.config.Servers[0].NotifyChannels = []string{"#alerts"}
				msg.Buffer = "#other"
				msg.Text = "ordinary message"
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

func TestSoundNotificationCmdRespectsCooldown(t *testing.T) {
	m := newActivityTestModel()

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time {
		return now
	}

	cmd := m.soundNotificationCmd()
	if cmd == nil {
		t.Fatal("first soundNotificationCmd() returned nil")
	}

	result := cmd()
	msg, ok := result.(tea.RawMsg)
	if !ok {
		t.Fatalf("soundNotificationCmd() returned %T, want tea.RawMsg", result)
	}
	if got, want := msg.Msg, "\a"; got != want {
		t.Fatalf("RawMsg.Msg = %q, want %q", got, want)
	}

	if cmd := m.soundNotificationCmd(); cmd != nil {
		t.Fatal("soundNotificationCmd() returned a command during cooldown")
	}

	now = now.Add(m.config.Notifications.Cooldown)
	if cmd := m.soundNotificationCmd(); cmd == nil {
		t.Fatal("soundNotificationCmd() returned nil after cooldown")
	}
}

func TestProcessIncomingLiveMentionReturnsBell(t *testing.T) {
	m := newActivityTestModel()

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time {
		return now
	}

	cmd := m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: now,
		From:      "alice",
		Text:      "dexuser: ping",
		Type:      irc.MessageTypeNormal,
	})

	assertBellCommand(t, cmd)
}

func TestProcessIncomingLiveDirectMessageReturnsBell(t *testing.T) {
	m := newActivityTestModel()

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time {
		return now
	}

	_, _ = m.getOrCreateBuffer("libera", "alice")
	cmd := m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:        "libera",
		Buffer:        "alice",
		DirectMessage: true,
		Timestamp:     now,
		From:          "alice",
		Text:          "hello",
		Type:          irc.MessageTypeNormal,
	})

	assertBellCommand(t, cmd)
}

func TestProcessIncomingOldMentionDoesNotReturnBell(t *testing.T) {
	m := newActivityTestModel()

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time {
		return now
	}

	cmd := m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: now.Add(-maxNotificationAge - time.Second),
		From:      "alice",
		Text:      "dexuser: old ping",
		Type:      irc.MessageTypeNormal,
	})

	if cmd != nil {
		t.Fatal("processIncomingMessage() returned a command for old playback mention")
	}
}

func assertBellCommand(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if !commandContainsBell(cmd) {
		t.Fatal("expected command batch to contain a terminal bell")
	}
}

func commandContainsBell(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}

	switch msg := cmd().(type) {
	case tea.RawMsg:
		return msg.Msg == "\a"
	case tea.BatchMsg:
		for _, batchedCmd := range msg {
			if commandContainsBell(batchedCmd) {
				return true
			}
		}
	}
	return false
}

func TestOpenHelpFromPalette(t *testing.T) {
	m := newActivityTestModel()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	// Open palette
	_, _ = m.Update(tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
	if !m.palette.IsVisible() {
		t.Fatal("palette should be open after ctrl+o")
	}

	// Trigger OpenHelpMsg
	_, _ = m.Update(palette.OpenHelpMsg{})
	if m.palette.IsVisible() {
		t.Fatal("palette should be closed after opening help")
	}
	if !m.help.IsVisible() {
		t.Fatal("help modal should be visible")
	}
}

func TestOpenHelpFromSlashCommand(t *testing.T) {
	m := newActivityTestModel()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	buf := m.getActiveBuffer()
	m.handleCommand(buf, commands.Command{Name: "help"})

	if !m.help.IsVisible() {
		t.Fatal("help modal should be visible after /help command")
	}
}

func TestHelpModalKeyHandlingAndClosing(t *testing.T) {
	m := newActivityTestModel()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	m.showHelp()
	if !m.help.IsVisible() {
		t.Fatal("help modal should be visible after showHelp()")
	}

	// While help is visible, view contains help content
	view := plainText(m.View().Content)
	if !strings.Contains(view, "Dex Keybindings") {
		t.Fatalf("UI view should contain Help modal overlay:\n%s", view)
	}

	// Esc closes help modal
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.help.IsVisible() {
		t.Fatal("help modal should be closed after pressing Esc")
	}

	// Reopen help and test 'q' key closing without leaking into chat input
	m.showHelp()
	if !m.help.IsVisible() {
		t.Fatal("help modal should be visible after second showHelp()")
	}

	_, _ = m.Update(tea.KeyPressMsg{Text: "q"})
	if m.help.IsVisible() {
		t.Fatal("help modal should be closed after pressing q")
	}

	buf := m.getActiveBuffer()
	chatView := plainText(buf.Chat.View())
	if strings.Contains(chatView, "q") && !strings.Contains(chatView, "Send message...") {
		t.Fatalf("chat input should not contain 'q' after closing help with q:\n%s", chatView)
	}
}

func TestViewReportsFocus(t *testing.T) {
	m := newActivityTestModel()
	view := m.View()
	if !view.ReportFocus {
		t.Fatal("expected View.ReportFocus to be true")
	}
}

func TestFocusAndBlurUpdatesTerminalFocusState(t *testing.T) {
	m := newActivityTestModel()
	if !m.terminalFocused {
		t.Fatal("expected terminalFocused to default to true")
	}

	_, _ = m.Update(tea.BlurMsg{})
	if m.terminalFocused {
		t.Fatal("expected terminalFocused to be false after BlurMsg")
	}

	_, _ = m.Update(tea.FocusMsg{})
	if !m.terminalFocused {
		t.Fatal("expected terminalFocused to be true after FocusMsg")
	}
}

func TestActiveBufferMessageWhenBlurredTriggersNotificationAndActivity(t *testing.T) {
	m := newActivityTestModel()
	key := makeBufferKey("libera", "#random")
	m.activeBuffer = key

	// Blur the terminal window
	_, _ = m.Update(tea.BlurMsg{})

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return now }

	cmd := m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: now,
		From:      "alice",
		Text:      "dexuser: ping",
		Type:      irc.MessageTypeNormal,
	})

	assertBellCommand(t, cmd)

	buf := m.buffers[key]
	if got := buf.UnreadCount; got != 1 {
		t.Fatalf("UnreadCount = %d, want 1 when terminal is blurred", got)
	}
	if got := buf.MentionCount; got != 1 {
		t.Fatalf("MentionCount = %d, want 1 when terminal is blurred", got)
	}

	// Refocusing terminal clears active buffer unread & mention counts
	_, _ = m.Update(tea.FocusMsg{})
	if got := buf.UnreadCount; got != 0 {
		t.Fatalf("UnreadCount after FocusMsg = %d, want 0", got)
	}
	if got := buf.MentionCount; got != 0 {
		t.Fatalf("MentionCount after FocusMsg = %d, want 0", got)
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
