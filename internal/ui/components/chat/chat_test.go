package chat

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type stubChannelMembers struct {
	nicks []string
}

func (s stubChannelMembers) HasUser(nick string) bool         { return false }
func (s stubChannelMembers) GetUserPrefix(nick string) string { return "" }
func (s stubChannelMembers) Nicknames() []string              { return s.nicks }

func TestTabCompletesAndCyclesMatchingNicknames(t *testing.T) {
	m := New(styles.RosePineTheme(), styles.NewUsernameColors(nil))
	m.SetChannelMembers(stubChannelMembers{nicks: []string{"Alice", "alex", "bob"}})
	m.input.SetValue("al")
	m.input.CursorEnd()

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := m.input.Value(); got != "Alice: " {
		t.Fatalf("expected first matching nick, got %q", got)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := m.input.Value(); got != "alex: " {
		t.Fatalf("expected Tab to cycle to next matching nick, got %q", got)
	}
}

func TestTabCompletesNicknameAtCursorWithoutAddressSuffix(t *testing.T) {
	m := New(styles.RosePineTheme(), styles.NewUsernameColors(nil))
	m.SetChannelMembers(stubChannelMembers{nicks: []string{"Alice"}})
	m.input.SetValue("hello al there")
	m.input.SetCursor(len([]rune("hello al")))

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := m.input.Value(); got != "hello Alice there" {
		t.Fatalf("expected nick at cursor to be completed in place, got %q", got)
	}
}

func TestFocusedInputAllowsMouseWheelToScrollViewport(t *testing.T) {
	m := newScrollableChatModel()

	before := m.viewport.YOffset()
	if before == 0 {
		t.Fatal("expected test setup to start at bottom with scrollable content")
	}

	var cmd tea.Cmd
	m, cmd = m.Update(tea.MouseWheelMsg{
		Button: tea.MouseWheelUp,
	})
	if cmd != nil {
		_ = cmd()
	}

	if m.viewport.YOffset() >= before {
		t.Fatalf("expected mouse wheel up to scroll while input is focused, offset before=%d after=%d", before, m.viewport.YOffset())
	}
}

func TestUpdateContentPreservesScrollPositionWhenNotAtBottom(t *testing.T) {
	m := newScrollableChatModel()
	m.viewport.ScrollUp(3)
	before := m.viewport.YOffset()
	if m.viewport.AtBottom() {
		t.Fatal("expected test setup to be scrolled away from bottom")
	}

	m.AddMessage(Message{
		Timestamp: time.Now(),
		Username:  "alice",
		Text:      "one more message",
	})

	if m.viewport.YOffset() != before {
		t.Fatalf("expected adding content to preserve scroll offset when not at bottom, before=%d after=%d", before, m.viewport.YOffset())
	}
}

func TestUpdateContentFollowsBottomWhenAlreadyAtBottom(t *testing.T) {
	m := newScrollableChatModel()
	if !m.viewport.AtBottom() {
		t.Fatal("expected test setup to start at bottom")
	}

	m.AddMessage(Message{
		Timestamp: time.Now(),
		Username:  "alice",
		Text:      "one more message",
	})

	if !m.viewport.AtBottom() {
		t.Fatalf("expected adding content to keep viewport at bottom, offset=%d", m.viewport.YOffset())
	}
}

func newScrollableChatModel() Model {
	theme := styles.RosePineTheme()
	m := New(theme, styles.NewUsernameColors(theme.Colors.Nicknames))
	m.SetNickname("alice")
	m.SetSize(80, 12)

	now := time.Now()
	for i := 0; i < 20; i++ {
		m.AddMessage(Message{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Username:  "alice",
			Text:      fmt.Sprintf("message %02d", i),
		})
	}

	return m
}
