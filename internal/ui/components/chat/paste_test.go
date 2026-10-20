package chat

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/styles"
)

func newPasteTestModel(width int) Model {
	theme := styles.RosePineTheme()
	m := New(theme, styles.NewUsernameColors(theme.Colors.Nicknames))
	m.SetNickname("alice")
	m.SetSize(width, 12)
	return m
}

func TestSplitPastedLines(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "single line", text: "hello", want: []string{"hello"}},
		{name: "trailing newline", text: "hello\n", want: []string{"hello"}},
		{name: "CRLF and CR", text: "a\r\nb\rc", want: []string{"a", "b", "c"}},
		{name: "drops blank lines", text: "a\n\n  \nb\n", want: []string{"a", "b"}},
		{name: "keeps indentation", text: "func f() {\n\treturn\n}", want: []string{"func f() {", "\treturn", "}"}},
		{name: "only blank lines", text: "\n \n", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitPastedLines(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitPastedLines(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestSingleLinePasteIsInsertedIntoInput(t *testing.T) {
	m := newPasteTestModel(80)
	m.SetInputValue("say: ")

	m, action, _ := m.Update(tea.PasteMsg{Content: "hello\n"})
	if action != nil || len(m.pastedLines) != 0 {
		t.Fatalf("single-line paste was held as lines: action = %#v, lines = %q", action, m.pastedLines)
	}
	if got := m.InputValue(); got != "say: hello" {
		t.Fatalf("input = %q, want the line inserted without its newline", got)
	}
}

func TestMultiLinePasteWaitsForEnterAndKeepsDraft(t *testing.T) {
	m := newPasteTestModel(80)
	m.SetInputValue("draft")

	m, action, _ := m.Update(tea.PasteMsg{Content: "one\ntwo\n"})
	if action != nil {
		t.Fatalf("paste sent without Enter: %#v", action)
	}
	if view := m.renderInputBox(); !strings.Contains(view, "2 pasted lines") || strings.Contains(view, "draft") {
		t.Fatalf("input row = %q, want the paste summary instead of the draft", view)
	}

	// Typing must not change the draft while it is hidden.
	m, _, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m, _, _ = m.Update(tea.PasteMsg{Content: "three"})

	m, action, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if action == nil || action.Text != "" || !reflect.DeepEqual(action.Pasted, []string{"one", "two", "three"}) {
		t.Fatalf("action = %#v, want the pasted lines", action)
	}
	if got := m.InputValue(); got != "draft" {
		t.Fatalf("input = %q, want the draft back", got)
	}
	if view := m.renderInputBox(); strings.Contains(view, "pasted lines") {
		t.Fatal("paste summary still shown after sending")
	}
}

func TestEscDropsMultiLinePaste(t *testing.T) {
	m := newPasteTestModel(80)
	m, _, _ = m.Update(tea.PasteMsg{Content: "one\ntwo"})

	m, action, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if action != nil || len(m.pastedLines) != 0 {
		t.Fatalf("Esc kept the paste: action = %#v, lines = %q", action, m.pastedLines)
	}
	if _, action, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); action != nil {
		t.Fatalf("Enter after Esc sent %#v", action)
	}
}

func TestPasteSummaryKeepsInputBoxSize(t *testing.T) {
	for _, width := range []int{80, 30} {
		m := newPasteTestModel(width)
		before := m.renderInputBox()
		m, _, _ = m.Update(tea.PasteMsg{Content: "one\ntwo\nthree"})
		after := m.renderInputBox()

		if lipgloss.Height(after) != lipgloss.Height(before) || lipgloss.Width(after) != lipgloss.Width(before) {
			t.Fatalf("width %d: input box is %dx%d with a paste, %dx%d without",
				width, lipgloss.Width(after), lipgloss.Height(after), lipgloss.Width(before), lipgloss.Height(before))
		}
	}
}
