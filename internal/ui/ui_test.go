package ui

import (
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/commands"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/history"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/components/chat"
	"github.com/dexchat/dex/internal/ui/components/palette"
	"github.com/dexchat/dex/internal/ui/components/users"
	"github.com/dexchat/dex/internal/ui/styles"
)

var errTestPersistence = errors.New("test persistence failure")

// updateIRC applies events as one batch pulled from server's queue.
func updateIRC(m *Model, server string, events ...irc.Event) tea.Cmd {
	_, cmd := m.Update(ircEventsMsg{server: server, events: events})
	return cmd
}

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

	_ = updateIRC(m, "libera", irc.ChannelPartedMsg{Server: "libera", Channel: "#go"})

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

// sidebarLine returns the rendered sidebar row for name.
func sidebarLine(m *Model, name string) string {
	for _, line := range strings.Split(plainText(m.channels.View(channelsPanelMaxWidth, 20)), "\n") {
		if strings.Contains(line, name) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func TestUnreadBadgeForChannelJoinedInSameBatch(t *testing.T) {
	m := newActivityTestModel()

	// A bouncer joins its channels and replays their buffers in one burst,
	// so the JOIN and the first unread message arrive in the same batch.
	_ = updateIRC(m, "libera",
		irc.ChannelJoinedMsg{Server: "libera", Channel: "#ai"},
		irc.BufferNewMessageMsg{
			Server:    "libera",
			Buffer:    "#ai",
			Timestamp: time.Now(),
			From:      "alice",
			Text:      "unread before close",
		},
	)

	if got := m.buffers[makeBufferKey("libera", "#ai")].UnreadCount; got != 1 {
		t.Fatalf("UnreadCount = %d, want 1", got)
	}
	if line := sidebarLine(m, "#ai"); !strings.HasSuffix(strings.TrimRight(line, "│ "), "1") {
		t.Fatalf("sidebar row = %q, want an unread badge of 1", line)
	}
}

func TestUnreadBadgeForNewDirectMessage(t *testing.T) {
	m := newActivityTestModel()

	_ = updateIRC(m, "libera", irc.BufferNewMessageMsg{
		Server:        "libera",
		Buffer:        "alice",
		DirectMessage: true,
		Timestamp:     time.Now(),
		From:          "alice",
		Text:          "hi",
	})

	if line := sidebarLine(m, "alice"); !strings.HasSuffix(strings.TrimRight(line, "│ "), "1") {
		t.Fatalf("sidebar row = %q, want an unread badge of 1", line)
	}
}

func directMessageFrom(from, text string) irc.BufferNewMessageMsg {
	return irc.BufferNewMessageMsg{
		Server:        "libera",
		Buffer:        from,
		DirectMessage: true,
		Timestamp:     time.Now(),
		From:          from,
		Text:          text,
	}
}

func TestDirectMessageIsHighlightedLikeMention(t *testing.T) {
	m := newActivityTestModel()

	_ = updateIRC(m, "libera", directMessageFrom("alice", "are you there?"))

	buf := m.buffers[makeBufferKey("libera", "alice")]
	if buf.MentionCount != 1 || buf.UnreadCount != 1 {
		t.Fatalf("MentionCount = %d, UnreadCount = %d, want 1 and 1", buf.MentionCount, buf.UnreadCount)
	}
	if line := sidebarLine(m, "alice"); !strings.HasSuffix(strings.TrimRight(line, "│ "), "@1") {
		t.Fatalf("sidebar row = %q, want a mention badge of @1", line)
	}
}

func TestIgnoredDirectMessageStaysPlainUnread(t *testing.T) {
	m := newActivityTestModel()
	m.config.Servers[0].IgnoreDirectMessagesFrom = []string{"alertbot"}

	_ = updateIRC(m, "libera", directMessageFrom("AlertBot", "disk is 91% full"))

	buf := m.buffers[makeBufferKey("libera", "AlertBot")]
	if buf.MentionCount != 0 || buf.UnreadCount != 1 {
		t.Fatalf("MentionCount = %d, UnreadCount = %d, want 0 and 1", buf.MentionCount, buf.UnreadCount)
	}
}

func TestOwnDirectMessageFromAnotherClientIsNotHighlighted(t *testing.T) {
	m := newActivityTestModel()
	msg := directMessageFrom("dexuser", "sent from my phone")
	msg.Buffer = "alice"

	_ = updateIRC(m, "libera", msg)

	if got := m.buffers[makeBufferKey("libera", "alice")].MentionCount; got != 0 {
		t.Fatalf("MentionCount = %d, want 0 for our own message", got)
	}
}

func TestUserListAfterSelfPartDoesNotRecreateChannel(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")

	_ = updateIRC(m, "libera",
		irc.ChannelPartedMsg{Server: "libera", Channel: "#go"},
		irc.UserListMsg{Server: "libera", Channel: "#go", Users: []string{"alice"}},
	)

	if _, exists := m.buffers[channelKey]; exists {
		t.Fatal("a user list after self-PART should not recreate the channel")
	}
	if content := plainText(m.channels.View(channelsPanelMaxWidth, 20)); strings.Contains(content, "#go") {
		t.Fatalf("a user list after self-PART should not restore the sidebar entry:\n%s", content)
	}
}

func TestSelfPartQueuesDirtyHistoryForPersistence(t *testing.T) {
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

	cmd := updateIRC(m, "libera", irc.ChannelPartedMsg{Server: "libera", Channel: "#go"})
	if cmd == nil {
		t.Fatal("dirty self-PART should schedule persistence")
	}
	if _, exists := m.buffers[key]; exists {
		t.Fatal("server-confirmed PART should remove the channel immediately")
	}
	if _, pending := m.persistence.detachedHistory[key]; !pending {
		t.Fatal("removed channel history should remain queued for persistence")
	}

	_, _ = m.Update(cmd())
	if _, pending := m.persistence.detachedHistory[key]; pending {
		t.Fatal("persisted channel history should leave the queue")
	}
}

func TestPersistenceMergesUnloadedHistory(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeTestHistory(t, "libera", "#go", "old message")
	m := newActivityTestModel()
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "new message",
	})

	cmd := m.startPersistence()
	if cmd == nil {
		t.Fatal("dirty unloaded history should schedule persistence")
	}
	_, _ = m.Update(cmd())

	entries := mustLoadHistory(t, "libera", "#go").Entries()
	if len(entries) != 2 {
		t.Fatalf("saved history entries = %d, want 2", len(entries))
	}
}

func TestSelfPartKeepsHistoryQueuedWhenPersistenceFails(t *testing.T) {
	m := newActivityTestModel()
	key := makeBufferKey("libera", "#go")
	log := history.NewLog()
	log.Insert(history.LogEntry{ServerTime: 1})
	m.persistence.detachedHistory[key] = detachedHistory{
		server: "libera",
		buffer: "#go",
		log:    log,
	}
	m.persistence.flushInFlight = true

	m.applyPersistenceResult(historyFlushFinishedMsg{
		results: []historyFlushResult{{key: key, log: log, revision: 1, err: errTestPersistence}},
	})

	if _, pending := m.persistence.detachedHistory[key]; !pending {
		t.Fatal("failed channel history should remain queued for retry")
	}
}

func TestLeaveDoesNotRemoveChannelBeforeServerConfirmation(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = submitChat(m, "/leave")

	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("/leave should keep the channel until the server confirms PART")
	}
}

func TestPartIsAliasForLeave(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = submitChat(m, "/part")

	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("/part should keep the channel until server confirmation")
	}
}

func TestCloseRemovesPrivateBufferAndSelectsServer(t *testing.T) {
	m := newActivityTestModel()
	m.getOrCreateBuffer("libera", "alice")
	m.directMessages.Add("libera", "alice")
	m.activeBuffer = makeBufferKey("libera", "alice")

	_, _ = submitChat(m, "/close")

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

func TestCloseResizesServerBufferAfterStaleWindowSize(t *testing.T) {
	m := newActivityTestModel()
	m.getOrCreateBuffer("libera", "alice")
	m.directMessages.Add("libera", "alice")

	// Resizing while the DM is active only resizes the DM's chat (per
	// WindowSizeMsg handling), leaving the server buffer's chat sized from
	// whatever it last had (its zero value here, since it was never active).
	m.activeBuffer = makeBufferKey("libera", "alice")
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	_, _ = submitChat(m, "/close")

	if got, want := m.activeBuffer, makeBufferKey("libera", ""); got != want {
		t.Fatalf("activeBuffer = %q, want %q", got, want)
	}

	// A chat freshly sized to the model's current dimensions is the ground
	// truth for what the (now active) server buffer should look like.
	want := chat.New(m.theme, m.usernameColors)
	want.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
	wantView := want.View()

	got := m.getActiveBuffer().Chat.View()
	if lipgloss.Height(got) != lipgloss.Height(wantView) {
		t.Fatalf("server buffer chat height after /close = %d, want %d", lipgloss.Height(got), lipgloss.Height(wantView))
	}
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(wantView, "\n")
	for i := range gotLines {
		if lipgloss.Width(gotLines[i]) != lipgloss.Width(wantLines[i]) {
			t.Fatalf("server buffer chat line %d width after /close = %d, want %d", i, lipgloss.Width(gotLines[i]), lipgloss.Width(wantLines[i]))
		}
	}
}

func TestChatCloseIsHandledBeforeBufferSwitch(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	private := m.getOrCreateBuffer("libera", "Alice")
	m.activeBuffer = private.Key
	_, _ = submitChat(m, "/close")
	if m.buffers[private.Key] != nil {
		t.Fatal("close should be handled synchronously during Enter")
	}
	m.activeBuffer = makeBufferKey("libera", "#random")
	if m.getActiveBuffer().Buffer != "#random" {
		t.Fatal("buffer switch failed")
	}
}

func TestChannelCloseCannotCloseAnotherConversation(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	private := m.getOrCreateBuffer("libera", "Alice")
	m.activeBuffer = makeBufferKey("libera", "#go")
	_, cmd := submitChat(m, "/close")
	m.activeBuffer = private.Key
	if cmd != nil {
		t.Fatal("invalid close must not schedule a delayed send")
	}
	if m.buffers[private.Key] == nil {
		t.Fatal("unrelated conversation was closed")
	}
}

func TestCloseQueuesDirtyPrivateHistoryForPersistence(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	key := makeBufferKey("libera", "alice")
	m.getOrCreateBuffer("libera", "alice")
	m.activeBuffer = key
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "alice",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello",
	})

	_, cmd := submitChat(m, "/close")
	if cmd == nil {
		t.Fatal("dirty /close should schedule persistence")
	}
	if _, exists := m.buffers[key]; exists {
		t.Fatal("/close should remove the private buffer immediately")
	}
	if _, pending := m.persistence.detachedHistory[key]; !pending {
		t.Fatal("closed private history should remain queued for persistence")
	}

	_, _ = m.Update(cmd())
	if _, pending := m.persistence.detachedHistory[key]; pending {
		t.Fatal("persisted private history should leave the queue")
	}
}

func TestCloseKeepsHistoryQueuedWhenPersistenceFails(t *testing.T) {
	m := newActivityTestModel()
	key := makeBufferKey("libera", "alice")
	log := history.NewLog()
	log.Insert(history.LogEntry{ServerTime: 1})
	m.persistence.detachedHistory[key] = detachedHistory{
		server: "libera",
		buffer: "alice",
		log:    log,
	}
	m.persistence.flushInFlight = true

	m.applyPersistenceResult(historyFlushFinishedMsg{
		results: []historyFlushResult{{key: key, log: log, revision: 1, err: errTestPersistence}},
	})

	if _, pending := m.persistence.detachedHistory[key]; !pending {
		t.Fatal("failed private history should remain queued for retry")
	}
}

func TestReopenedBufferKeepsDetachedHistoryWhilePersistenceRuns(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	key := makeBufferKey("libera", "alice")
	buffer := m.getOrCreateBuffer("libera", "alice")
	m.activeBuffer = key
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "alice",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "first",
	})

	firstFlush := m.handleCloseCommand(buffer, nil)
	if firstFlush == nil {
		t.Fatal("first close should start persistence")
	}
	reopened := m.getOrCreateBuffer("libera", "alice")
	if reopened.History != buffer.History {
		t.Fatal("reopened buffer should reuse history still being persisted")
	}
	m.activeBuffer = key
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "alice",
		Timestamp: time.Now().Add(time.Second),
		From:      "alice",
		Text:      "second",
	})
	if cmd := m.handleCloseCommand(reopened, nil); cmd != nil {
		t.Fatal("second close should queue behind the in-flight persistence")
	}

	_, secondFlush := m.Update(firstFlush())
	if secondFlush == nil {
		t.Fatal("newer detached history should trigger a second flush")
	}
	_, _ = m.Update(secondFlush())
	if _, pending := m.persistence.detachedHistory[key]; pending {
		t.Fatal("latest detached history should leave the queue after persistence")
	}
	if got := len(mustLoadHistory(t, "libera", "alice").Entries()); got != 2 {
		t.Fatalf("persisted history contains %d messages, want 2", got)
	}
}

func TestShutdownWaitsForInFlightPersistenceAndFlushesNewRevision(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "first",
	})

	firstFlush := m.startPersistence()
	if firstFlush == nil {
		t.Fatal("dirty history should start persistence")
	}
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now().Add(time.Second),
		From:      "alice",
		Text:      "second",
	})
	m.persistence.shutdownRequested = true
	if cmd := m.startPersistence(); cmd != nil {
		t.Fatal("shutdown should wait for the in-flight command")
	}

	_, secondFlush := m.Update(firstFlush())
	if secondFlush == nil {
		t.Fatal("shutdown should persist changes newer than the first snapshot")
	}
	_, quitCmd := m.Update(secondFlush())
	if quitCmd == nil {
		t.Fatal("shutdown should quit after the final snapshot is persisted")
	}
}

func TestCtrlCQuitsAfterPersistenceCompletes(t *testing.T) {
	defer func() { time.Sleep(1600 * time.Millisecond) }()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m := newActivityTestModel()
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello",
	})

	_, _ = m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("second Ctrl+C should schedule the final persistence")
	}
	if m.persistence.shutdownRequested == false {
		t.Fatal("second Ctrl+C should request shutdown")
	}

	model, quitCmd := m.Update(cmd())
	if model != m {
		t.Fatal("persistence completion should keep the same model")
	}
	if quitCmd == nil {
		t.Fatal("persistence completion should return tea.Quit")
	}
}

func TestShutdownPersistenceFailureKeepsAppOpenForRetry(t *testing.T) {
	m := newActivityTestModel()
	m.getActiveBuffer().Chat.SetSize(80, 10)
	m.readStateDirty = true
	m.persistence.flushInFlight = true
	m.persistence.flushPending = true
	m.persistence.shutdownRequested = true

	cmd := m.applyPersistenceResult(historyFlushFinishedMsg{
		readDirty:    true,
		readStateErr: errTestPersistence,
	})

	if cmd != nil {
		t.Fatal("failed shutdown persistence should keep the app running")
	}
	if m.persistence.shutdownRequested {
		t.Fatal("failed shutdown persistence should cancel the shutdown request")
	}
	if m.persistence.flushPending {
		t.Fatal("failed shutdown persistence should wait for an explicit retry")
	}
	if !m.readStateDirty {
		t.Fatal("failed read-state persistence should remain dirty")
	}
	if view := plainText(m.getActiveBuffer().Chat.View()); !strings.Contains(view, "could not save local data") || !strings.Contains(view, "press ctrl+c to retry") {
		t.Fatalf("active buffer does not explain the persistence failure:\n%s", view)
	}
	if retry := m.startPersistence(); retry == nil {
		t.Fatal("dirty state should be available for a later persistence retry")
	}
}

func TestCloseDoesNotCloseChannel(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	m.activeBuffer = channelKey

	_, _ = submitChat(m, "/close")

	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("/close should not remove a channel buffer")
	}
}

func TestJoinDoesNotCreateChannelBeforeServerConfirmation(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#new")

	_, _ = submitChat(m, "/join #new")

	if _, exists := m.buffers[channelKey]; exists {
		t.Fatal("/join should not create a channel before the server confirms JOIN")
	}

	_ = updateIRC(m, "libera", irc.ChannelJoinedMsg{Server: "libera", Channel: "#new"})
	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("self-JOIN confirmation should create the channel buffer")
	}
}

func TestJoinWithoutChannelShowsUsageInsteadOfCreatingBuffer(t *testing.T) {
	m := newActivityTestModel()
	active := m.getActiveBuffer()
	active.Chat.SetSize(80, 10)

	_, _ = submitChat(m, "/join")

	if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /join <channel> [key]") {
		t.Fatalf("missing /join usage error:\n%s", view)
	}
}

func TestIRCCommandErrorIsShownInIssuingBuffer(t *testing.T) {
	m := newActivityTestModel()
	channel := m.buffers[makeBufferKey("libera", "#go")]
	channel.Chat.SetSize(80, 10)
	m.activeBuffer = makeBufferKey("libera", "#random")
	active := m.getActiveBuffer()
	active.Chat.SetSize(80, 10)

	msg := ircCommand(channel, func() error { return errors.New("irc: not in channel #go") })()
	_, _ = m.Update(msg)

	if view := plainText(channel.Chat.View()); !strings.Contains(view, "irc: not in channel #go") {
		t.Fatalf("command error missing from the issuing buffer:\n%s", view)
	}
	if view := plainText(active.Chat.View()); strings.Contains(view, "not in channel") {
		t.Fatalf("command error leaked into the active buffer:\n%s", view)
	}
}

func TestIRCCommandErrorFallsBackToServerBufferAfterClose(t *testing.T) {
	m := newActivityTestModel()
	server := m.buffers[makeBufferKey("libera", "")]
	server.Chat.SetSize(80, 10)
	closed := &Buffer{Key: makeBufferKey("libera", "alice"), Server: "libera", Buffer: "alice"}

	msg := ircCommand(closed, func() error { return errors.New("irc: not connected to libera") })()
	_, _ = m.Update(msg)

	if view := plainText(server.Chat.View()); !strings.Contains(view, "irc: not connected to libera") {
		t.Fatalf("command error for a closed buffer missing from the server buffer:\n%s", view)
	}
}

func TestSuccessfulIRCCommandProducesNoMessage(t *testing.T) {
	buf := &Buffer{Key: makeBufferKey("libera", "#go"), Server: "libera", Buffer: "#go"}
	if msg := ircCommand(buf, func() error { return nil })(); msg != nil {
		t.Fatalf("successful command returned %#v, want nil", msg)
	}
}

func TestListWithTooManyArgumentsShowsUsage(t *testing.T) {
	m := newActivityTestModel()
	active := m.getActiveBuffer()
	active.Chat.SetSize(80, 10)

	_, _ = submitChat(m, "/list #go #rust")

	if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /list [channel]") {
		t.Fatalf("missing /list usage error:\n%s", view)
	}
}

func TestMsgCreatesAndSelectsPrivateBuffer(t *testing.T) {
	m := newActivityTestModel()
	serverBuffer := m.getActiveBuffer()

	_, _ = submitChat(m, "/msg Alice hello there")

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

	_, _ = submitChat(m, "/msg Alice")

	if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /msg <user> <message>") {
		t.Fatalf("missing /msg usage error:\n%s", view)
	}
}

func TestWhoisWithoutNickOutsidePrivateMessageShowsUsage(t *testing.T) {
	for _, input := range []string{"/whois", "/whois alice bob"} {
		t.Run(input, func(t *testing.T) {
			m := newActivityTestModel()
			m.activeBuffer = makeBufferKey("libera", "#go")
			active := m.getActiveBuffer()
			active.Chat.SetSize(80, 10)

			_, _ = submitChat(m, input)

			if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /whois <nick>") {
				t.Fatalf("missing /whois usage error:\n%s", view)
			}
		})
	}
}

func TestNickRequiresExactlyOneNickname(t *testing.T) {
	for _, input := range []string{"/nick", "/nick alice bob"} {
		t.Run(input, func(t *testing.T) {
			m := newActivityTestModel()
			active := m.getActiveBuffer()
			active.Chat.SetSize(80, 10)

			_, _ = submitChat(m, input)

			if view := plainText(active.Chat.View()); !strings.Contains(view, "usage: /nick <nickname>") {
				t.Fatalf("missing /nick usage error:\n%s", view)
			}
		})
	}
}

func TestMeShowsOwnActionInChannel(t *testing.T) {
	m := newActivityTestModel()
	m.activeBuffer = makeBufferKey("libera", "#go")
	active := m.getActiveBuffer()
	active.Chat.SetNickname("dexuser")
	active.Chat.SetSize(80, 10)

	_, _ = submitChat(m, "/me waves at everyone")

	want := "* dexuser waves at everyone"
	if view := plainText(active.Chat.View()); !strings.Contains(view, want) {
		t.Fatalf("missing %q in channel:\n%s", want, view)
	}
}

func TestMeRejectsServerBufferAndEmptyAction(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "/me waves", want: "error: /me is only available in a channel or private message"},
		{input: "/me", want: "usage: /me <action>"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := newActivityTestModel()
			m.activeBuffer = makeBufferKey("libera", "")
			active := m.getActiveBuffer()
			active.Chat.SetSize(80, 10)

			_, _ = submitChat(m, tt.input)

			view := plainText(active.Chat.View())
			if !strings.Contains(view, tt.want) {
				t.Fatalf("missing %q:\n%s", tt.want, view)
			}
			if strings.Contains(view, "* ") {
				t.Fatalf("rejected /me was shown as an action:\n%s", view)
			}
		})
	}
}

func TestIncomingActionMentionIsHighlightedAndRendered(t *testing.T) {
	m := newActivityTestModel()
	m.activeBuffer = makeBufferKey("libera", "#random")
	channel := m.buffers[makeBufferKey("libera", "#go")]
	channel.Chat.SetNickname("dexuser")
	channel.Chat.SetSize(80, 10)

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "waves at dexuser",
		Action:    true,
	})

	if channel.MentionCount != 1 {
		t.Fatalf("MentionCount = %d, want 1", channel.MentionCount)
	}
	channel.Chat.FlushQueue()
	if view := plainText(channel.Chat.View()); !strings.Contains(view, "* alice waves at dexuser") {
		t.Fatalf("missing rendered action:\n%s", view)
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
	m := newActivityTestModel()
	events := make([]irc.Event, maxPlaybackMessagesPerUpdate+1)
	for i := range events {
		events[i] = irc.BufferNewMessageMsg{
			Server:    "libera",
			Buffer:    "#go",
			Timestamp: time.Unix(int64(i+1), 0),
			From:      "alice",
			Text:      "playback",
		}
	}

	_ = updateIRC(m, "libera", events...)
	if got := len(m.ircPending["libera"]); got != 1 {
		t.Fatalf("pending events after first update = %d, want 1", got)
	}

	_, _ = m.Update(ircContinueMsg{server: "libera"})
	if _, pending := m.ircPending["libera"]; pending {
		t.Fatalf("pending events after continuing = %d, want 0", len(m.ircPending["libera"]))
	}
	if got := len(m.buffers[makeBufferKey("libera", "#go")].History.Entries()); got != len(events) {
		t.Fatalf("applied %d playback messages, want %d", got, len(events))
	}
}

func TestSelfPartAfterSplitPlaybackDoesNotRecreateChannel(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	events := make([]irc.Event, 0, maxPlaybackMessagesPerUpdate+1)
	for i := range maxPlaybackMessagesPerUpdate {
		events = append(events, irc.BufferNewMessageMsg{
			Server:    "libera",
			Buffer:    "#go",
			Timestamp: time.Unix(int64(i+1), 0),
			From:      "alice",
			Text:      "before part",
		})
	}
	events = append(events, irc.ChannelPartedMsg{Server: "libera", Channel: "#go"})

	_ = updateIRC(m, "libera", events...)
	if _, exists := m.buffers[channelKey]; !exists {
		t.Fatal("PART was applied before the messages that preceded it")
	}

	_, _ = m.Update(ircContinueMsg{server: "libera"})
	if _, exists := m.buffers[channelKey]; exists {
		t.Fatal("self-PART at the end of a split batch should remove the channel")
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

func TestOwnEchoIsStoredWithoutBeingShownAgain(t *testing.T) {
	m := newActivityTestModel()
	m.activeBuffer = makeBufferKey("libera", "#go")
	channel := m.getActiveBuffer()
	channel.Chat.SetSize(80, 10)

	// The UI shows a sent message immediately, then receives its echo.
	channel.Chat.AddMessage(chat.Message{Timestamp: time.Now(), Username: "dexuser", Text: "sent once"})
	_ = updateIRC(m, "libera", irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "dexuser",
		Text:      "sent once",
		OwnEcho:   true,
	})
	channel.Chat.FlushQueue()

	if got := len(channel.History.Entries()); got != 1 {
		t.Fatalf("history has %d entries, want the echo stored once", got)
	}
	if got := strings.Count(plainText(channel.Chat.View()), "sent once"); got != 1 {
		t.Fatalf("message shown %d times, want once", got)
	}
}

func TestInactiveUserListRendersWhenBufferIsSelected(t *testing.T) {
	m := newActivityTestModel()
	_ = updateIRC(m, "libera", irc.UserListMsg{
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

func TestQuitInInactiveChannelDimsNicknameOnceSelected(t *testing.T) {
	m := newActivityTestModel()
	var cmds []tea.Cmd

	// alice is present in #random while it's the active buffer, and her
	// message is rendered with an active nickname color.
	m.selectBuffer("libera", "#random", &cmds)
	_ = updateIRC(m, "libera", irc.UserListMsg{
		Server:  "libera",
		Channel: "#random",
		Users:   []string{"alice"},
	})
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#random",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "hello from random",
	})
	random := m.buffers[makeBufferKey("libera", "#random")]
	random.Chat.FlushQueue()

	inactiveStyle := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Chat.InactiveNickname).
		Background(m.theme.Colors.Base.Background)
	inactiveNick := inactiveStyle.Render(" alice ")
	if strings.Contains(random.Chat.View(), inactiveNick) {
		t.Fatal("expected alice's nickname to render active while she's still in the channel")
	}

	// The user switches away from #random, and alice quits while it's no
	// longer the active buffer.
	m.selectBuffer("libera", "#go", &cmds)
	_ = updateIRC(m, "libera", irc.UserListMsg{
		Server:  "libera",
		Channel: "#random",
		Users:   []string{},
	})

	// Switching back to #random should show alice's already-rendered line
	// recolored as inactive, not the stale active color.
	m.selectBuffer("libera", "#random", &cmds)
	if !strings.Contains(random.Chat.View(), inactiveNick) {
		t.Fatal("expected alice's nickname to render dimmed after quitting in an inactive channel")
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

func TestNewUsesConfiguredTheme(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	tests := []struct {
		name       string
		configured string
		want       styles.Theme
	}{
		{name: "Ayu Dark", configured: config.ThemeAyuDark, want: styles.AyuDarkTheme()},
		{name: "Dracula", configured: config.ThemeDracula, want: styles.DraculaTheme()},
		{name: "Gruvbox Dark", configured: config.ThemeGruvboxDark, want: styles.GruvboxDarkTheme()},
		{name: "Solarized Light", configured: config.ThemeSolarizedLight, want: styles.SolarizedLightTheme()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(&config.Config{UI: config.UI{Theme: tt.configured}})
			gotR, gotG, gotB, gotA := m.theme.Colors.Base.Background.RGBA()
			wantR, wantG, wantB, wantA := tt.want.Colors.Base.Background.RGBA()
			if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
				t.Fatalf("theme background = %#v, want %#v", m.theme.Colors.Base.Background, tt.want.Colors.Base.Background)
			}
		})
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
	if err := history.FlushDirectMessagesSnapshot(*directMessages); err != nil {
		t.Fatalf("failed to persist direct messages: %v", err)
	}

	m := New(&config.Config{Servers: []*config.Server{
		{Name: "libera", Nickname: "dexuser"},
		{Name: "oftc", Nickname: "dexuser"},
	}})

	if _, ok := m.buffers[makeBufferKey("libera", "alice")]; ok {
		t.Fatal("persisted direct message restored before the server connected")
	}

	cmd := updateIRC(m, "libera", irc.BufferNewMessageMsg{
		Server: "libera",
		Type:   irc.MessageTypeConnected,
	})
	if cmd != nil {
		cmd()
	}

	if _, ok := m.buffers[makeBufferKey("libera", "alice")]; !ok {
		t.Fatal("persisted direct message was not restored after the server connected")
	}
	if _, ok := m.buffers[makeBufferKey("oftc", "bob")]; ok {
		t.Fatal("connecting one server restored another server's direct message")
	}

	cmd = updateIRC(m, "libera", irc.BufferNewMessageMsg{
		Server: "libera",
		Type:   irc.MessageTypeConnected,
	})
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
	buf := m.getOrCreateBuffer("libera", "#go")
	if got := len(buf.History.Entries()); got != 0 {
		t.Fatalf("new buffer loaded %d history entries synchronously, want 0", got)
	}
	if sidebarLine(m, "#go") == "" {
		t.Fatal("a new buffer should be in the sidebar immediately")
	}
	if buf.historyState == historyLoading {
		t.Fatal("discovered channel should not start disk history loading at startup")
	}

	m.Update(m.requestHistoryLoad(buf)())
	if got := len(buf.History.Entries()); got != 1 {
		t.Fatalf("loaded history entries = %d, want 1", got)
	}
}

func writeTestHistory(t *testing.T, server, buffer, text string) {
	t.Helper()

	entry := history.LogEntry{
		ReceivedAt: time.Now().UnixNano(),
		ServerTime: time.Now().UnixNano(),
		Username:   "alice",
		Text:       text,
	}
	if err := history.FlushSnapshot(server, buffer, []history.LogEntry{entry}); err != nil {
		t.Fatalf("failed to write test history: %v", err)
	}
}

// persistNow runs the app's persistence command to completion.
func persistNow(t *testing.T, m *Model) {
	t.Helper()
	cmd := m.startPersistence()
	if cmd == nil {
		t.Fatal("expected pending local data to persist")
	}
	_, _ = m.Update(cmd())
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
	persistNow(t, first)

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
	persistNow(t, m)

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

	// Run the shutdown the second ctrl+c starts: save, then quit.
	save := m.requestShutdown()
	if save == nil {
		t.Fatal("shutdown with unsaved data should persist before quitting")
	}
	_, quit := m.Update(save())
	if quit == nil {
		t.Fatal("shutdown did not quit after persisting")
	}
	if _, ok := quit().(tea.QuitMsg); !ok {
		t.Fatal("shutdown did not quit after persisting")
	}

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

	m.getOrCreateBuffer("libera", "alice")
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

func TestNickUpdateChangesOnlyMatchingServerWithoutRenderingInactiveBuffers(t *testing.T) {
	cfg := &config.Config{
		Servers: []*config.Server{
			{Name: "libera", Nickname: "old-libera", Channels: []string{"#go"}},
			{Name: "oftc", Nickname: "old-oftc", Channels: []string{"#rust"}},
		},
	}
	m := New(cfg)
	m.width = 100
	m.height = 20

	liberaChannel := m.buffers[makeBufferKey("libera", "#go")]
	liberaChannel.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
	liberaChannel.Chat.QueueMessage(chat.Message{Text: "deferred message"})

	_ = updateIRC(m, "libera", irc.NickUpdateMsg{Server: "libera", Nick: "new-libera"})

	if got, want := m.buffers[makeBufferKey("libera", "")].Chat.Nickname(), "new-libera"; got != want {
		t.Fatalf("libera server nickname = %q, want %q", got, want)
	}
	if got, want := liberaChannel.Chat.Nickname(), "new-libera"; got != want {
		t.Fatalf("libera channel nickname = %q, want %q", got, want)
	}
	if got, want := m.buffers[makeBufferKey("oftc", "")].Chat.Nickname(), "old-oftc"; got != want {
		t.Fatalf("oftc server nickname = %q, want %q", got, want)
	}
	if got, want := m.buffers[makeBufferKey("oftc", "#rust")].Chat.Nickname(), "old-oftc"; got != want {
		t.Fatalf("oftc channel nickname = %q, want %q", got, want)
	}
	if view := plainText(liberaChannel.Chat.View()); strings.Contains(view, "deferred message") {
		t.Fatalf("NickUpdateMsg rendered an inactive buffer prematurely:\n%s", view)
	}
}

func TestTopicUpdateDefersInactiveBufferRenderingUntilSelection(t *testing.T) {
	m := newActivityTestModel()
	channelKey := makeBufferKey("libera", "#go")
	channel := m.buffers[channelKey]
	channel.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
	channel.Chat.QueueMessage(chat.Message{Text: "deferred message"})

	_ = updateIRC(m, "libera", irc.ChannelTopicMsg{
		Server:  "libera",
		Channel: "#go",
		Topic:   "A deferred topic",
	})

	if view := plainText(channel.Chat.View()); strings.Contains(view, "deferred message") {
		t.Fatalf("ChannelTopicMsg rendered an inactive buffer prematurely:\n%s", view)
	}

	_, _ = m.Update(palette.ChannelSelectionMsg{Server: "libera", Channel: "#go"})
	view := plainText(channel.Chat.View())
	if !strings.Contains(view, "A deferred topic") {
		t.Fatalf("selected buffer is missing its updated topic:\n%s", view)
	}
	if !strings.Contains(view, "deferred message") {
		t.Fatalf("selected buffer is missing its queued message:\n%s", view)
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

func submitChat(m *Model, text string) (tea.Model, tea.Cmd) {
	m.getActiveBuffer().Chat.SetInputValue(text)
	return m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
}

func mustLoadHistory(t *testing.T, server, buffer string) *history.Log {
	t.Helper()
	log, err := history.Load(server, buffer)
	if err != nil {
		t.Fatalf("history.Load() error = %v", err)
	}
	return log
}

// testHistoryPath returns the single history file stored for server.
func testHistoryPath(t *testing.T, server string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(os.Getenv("XDG_DATA_HOME"), "dex", "history", server, "*.json.gz"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("history files for %s = %v (err %v), want exactly one", server, matches, err)
	}
	return matches[0]
}

// corruptTestHistory replaces a stored history with one readable entry
// followed by damaged data.
func corruptTestHistory(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	gz := gzip.NewWriter(file)
	_, _ = gz.Write([]byte(`{"server_time":1,"username":"alice","text":"readable"}` + "\n{broken"))
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip Close() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

// makeTestHistoryUnreadable removes read permission from the stored history.
func makeTestHistoryUnreadable(t *testing.T, path string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("file permissions are not enforced for root")
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
}

func TestCorruptHistoryIsQuarantinedAndReadableEntriesSaved(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeTestHistory(t, "libera", "#go", "stored message")
	path := testHistoryPath(t, "libera")
	corruptTestHistory(t, path)

	m := newActivityTestModel()
	buf := m.buffers[makeBufferKey("libera", "#go")]
	m.Update(m.requestHistoryLoad(buf)())

	if buf.historyState == historyUnreadable {
		t.Fatal("corrupt history should be recovered, not disable persistence")
	}
	if got := len(buf.History.Entries()); got != 1 {
		t.Fatalf("recovered entries = %d, want 1", got)
	}
	if corrupt, _ := filepath.Glob(path + ".corrupt-*"); len(corrupt) != 1 {
		t.Fatalf("quarantined files = %v, want one", corrupt)
	}

	cmd := m.startPersistence()
	if cmd == nil {
		t.Fatal("recovered history should be saved again")
	}
	m.Update(cmd())
	if entries := mustLoadHistory(t, "libera", "#go").Entries(); len(entries) != 1 || entries[0].Text != "readable" {
		t.Fatalf("saved history = %#v, want the readable entry", entries)
	}
}

func TestUnreadableHistoryIsNeverReplaced(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeTestHistory(t, "libera", "#go", "stored message")
	path := testHistoryPath(t, "libera")
	makeTestHistoryUnreadable(t, path)

	m := newActivityTestModel()
	buf := m.buffers[makeBufferKey("libera", "#go")]
	m.Update(m.requestHistoryLoad(buf)())
	if buf.historyState != historyUnreadable {
		t.Fatalf("history state = %v, want historyUnreadable", buf.historyState)
	}

	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "live message",
	})
	if cmd := m.startPersistence(); cmd != nil {
		t.Fatal("unreadable history must not be persisted")
	}

	_ = os.Chmod(path, 0o600)
	if entries := mustLoadHistory(t, "libera", "#go").Entries(); len(entries) != 1 || entries[0].Text != "stored message" {
		t.Fatalf("stored history = %#v, want the original entry", entries)
	}
}

func TestUnreadableHistoryDuringMergeDoesNotBlockShutdown(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeTestHistory(t, "libera", "#go", "stored message")
	path := testHistoryPath(t, "libera")
	makeTestHistoryUnreadable(t, path)

	m := newActivityTestModel()
	m.processIncomingMessage(irc.BufferNewMessageMsg{
		Server:    "libera",
		Buffer:    "#go",
		Timestamp: time.Now(),
		From:      "alice",
		Text:      "live message",
	})
	m.persistence.shutdownRequested = true
	cmd := m.startPersistence()
	if cmd == nil {
		t.Fatal("dirty history should schedule persistence")
	}

	_, next := m.Update(cmd())
	if next == nil {
		t.Fatal("shutdown should quit after skipping unreadable history")
	}
	if _, ok := next().(tea.QuitMsg); !ok {
		t.Fatalf("next command returned %T, want tea.QuitMsg", next())
	}
	if m.buffers[makeBufferKey("libera", "#go")].historyState != historyUnreadable {
		t.Fatal("buffer should stop persisting after a merge read error")
	}
}
