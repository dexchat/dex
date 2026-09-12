package help

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func TestHelpVisibility(t *testing.T) {
	m := New(styles.RosePineTheme())
	if m.IsVisible() {
		t.Fatal("help modal should not be visible initially")
	}

	m.Show()
	if !m.IsVisible() {
		t.Fatal("help modal should be visible after Show()")
	}

	m.Close()
	if m.IsVisible() {
		t.Fatal("help modal should not be visible after Close()")
	}

	m.Toggle()
	if !m.IsVisible() {
		t.Fatal("help modal should be visible after Toggle()")
	}

	m.Toggle()
	if m.IsVisible() {
		t.Fatal("help modal should not be visible after second Toggle()")
	}
}

func TestHelpClosesOnKeypress(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{name: "Esc key", msg: tea.KeyPressMsg{Code: tea.KeyEscape}},
		{name: "Enter key", msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
		{name: "q key", msg: tea.KeyPressMsg{Text: "q"}},
		{name: "Q key", msg: tea.KeyPressMsg{Text: "Q"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(styles.RosePineTheme())
			m.Show()

			m, _ = m.Update(tt.msg)
			if m.IsVisible() {
				t.Fatalf("help modal should be closed after pressing %s", tt.name)
			}
		})
	}
}

func TestHelpViewContainsShortcuts(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.SetSize(80, 24)

	// Invisible view returns empty string
	if view := m.View(); view != "" {
		t.Fatalf("invisible view returned %q, want empty string", view)
	}

	m.Show()
	view := m.View()
	if !strings.Contains(view, "Dex Keybindings") {
		t.Fatal("help view should contain title")
	}
	if !strings.Contains(view, "ctrl+o") {
		t.Fatal("help view should contain ctrl+o shortcut")
	}
	if !strings.Contains(view, "ctrl+g") {
		t.Fatal("help view should contain ctrl+g shortcut")
	}
	if !strings.Contains(view, "ctrl+6") {
		t.Fatal("help view should contain ctrl+6 shortcut")
	}
	if !strings.Contains(view, "pgup / pgdn") {
		t.Fatal("help view should contain chat scrolling shortcuts")
	}
	if !strings.Contains(view, "ctrl+x+e") || !strings.Contains(view, "edit in editor") {
		t.Fatal("help view should contain edit-in-editor shortcut")
	}
	if !strings.Contains(view, "ctrl+c") {
		t.Fatal("help view should contain ctrl+c shortcut")
	}

	if width := lipgloss.Width(view); width > 80 {
		t.Fatalf("help view width = %d, want at most 80", width)
	}
}
