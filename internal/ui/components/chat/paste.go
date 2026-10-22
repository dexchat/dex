package chat

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/dexchat/dex/internal/ui/components/keybindings"
)

// splitPastedLines returns the lines of pasted text. IRC has no empty
// messages, so blank lines are dropped; other lines keep their spacing.
func splitPastedLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	var lines []string
	for line := range strings.SplitSeq(text, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// readClipboard reads the clipboard for ctrl+v as textinput would, but
// returns it as a paste, because textinput joins its lines. Terminals that
// paste on ctrl+v themselves send a tea.PasteMsg instead. Like textinput, a
// clipboard that cannot be read pastes nothing.
func readClipboard() tea.Msg {
	text, err := clipboard.ReadAll()
	if err != nil {
		return nil
	}
	return tea.PasteMsg{Content: text}
}

// updatePastedLines handles keys while a multi-line paste waits. Typing is
// ignored so the draft underneath cannot change unseen.
func (m *Model) updatePastedLines(msg tea.KeyPressMsg, keyMap keybindings.KeyMap) (Model, *SendAction, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEnter:
		lines := m.pastedLines
		m.pastedLines = nil
		return *m, &SendAction{Pasted: lines}, nil
	case msg.Code == tea.KeyEscape:
		m.pastedLines = nil
	case key.Matches(msg, keyMap.ScrollUp, keyMap.ScrollDown):
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return *m, nil, cmd
	}
	return *m, nil, nil
}

// renderPastedLines renders the waiting paste in the input row, padded and
// cut to the width the input itself takes, which includes a cursor cell.
func (m *Model) renderPastedLines() string {
	width := m.input.Width() + 1
	summary := fmt.Sprintf("%d pasted lines: Enter to send, Esc to cancel", len(m.pastedLines))
	summary = lipgloss.NewStyle().Inline(true).MaxWidth(width).Render(summary)
	summary += strings.Repeat(" ", max(0, width-lipgloss.Width(summary)))
	return m.theme.Styles.InputField.
		Foreground(m.theme.Colors.Base.Accent).
		Render(summary)
}
