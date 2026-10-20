package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/irc"
)

// newPasteTestModel returns a model on #go whose manager is never
// connected, so every send fails.
func newPasteTestModel(t *testing.T) (*Model, *Buffer) {
	t.Helper()
	m := newActivityTestModel()
	m.SetManager(irc.NewClientManager([]*config.Server{{Name: "libera", Nickname: "dexuser"}}))
	m.activeBuffer = makeBufferKey("libera", "#go")
	channel := m.getActiveBuffer()
	channel.Chat.SetSize(80, 30)
	return m, channel
}

func pasteAndSend(m *Model, text string) tea.Cmd {
	m.Update(tea.PasteMsg{Content: text})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return cmd
}

func TestPastedLinesAreShownAndSentAsText(t *testing.T) {
	m, channel := newPasteTestModel(t)

	cmd := pasteAndSend(m, "/nonexistent one\n/join #two\n")
	if cmd == nil {
		t.Fatal("pasting returned no send command")
	}
	view := plainText(channel.Chat.View())
	if !strings.Contains(view, "/nonexistent one") || !strings.Contains(view, "/join #two") {
		t.Fatalf("pasted lines are not shown:\n%s", view)
	}
	if strings.Contains(view, "unknown command") {
		t.Fatalf("a pasted line was run as a command:\n%s", view)
	}
}

func TestPastedLinesReportHowManyWereNotSent(t *testing.T) {
	m, channel := newPasteTestModel(t)

	cmd := pasteAndSend(m, "one\ntwo\nthree")
	m.Update(cmd())

	if view := plainText(channel.Chat.View()); !strings.Contains(view, "not connected to libera; 3 pasted lines were not sent") {
		t.Fatalf("chat does not report the unsent lines:\n%s", view)
	}
}

func TestPasteDoesNotReachChatUnderPalette(t *testing.T) {
	m, channel := newPasteTestModel(t)
	m.palette.Toggle()
	m.Update(tea.PasteMsg{Content: "one\ntwo"})
	m.palette.Toggle()
	if strings.Contains(plainText(channel.Chat.View()), "pasted lines") {
		t.Fatal("paste into the palette was also held by the chat")
	}
}
