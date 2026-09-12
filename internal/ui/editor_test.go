package ui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/ui/components/palette"
)

func TestNormalizeEditorText(t *testing.T) {
	if got, want := normalizeEditorText("first\r\nsecond\n"), "first second"; got != want {
		t.Fatalf("normalizeEditorText() = %q, want %q", got, want)
	}
}

func TestFinishEditingInputUpdatesOriginalBufferAndRemovesFile(t *testing.T) {
	m := newActivityTestModel()
	buffer := m.activeBuffer
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("edited\nmessage\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	m.finishEditingInput(editorFinishedMsg{buffer: buffer, path: path})

	if got, want := m.buffers[buffer].Chat.InputValue(), "edited message"; got != want {
		t.Fatalf("input value = %q, want %q", got, want)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists: %v", err)
	}
}

func TestFinishEditingInputPreservesDraftWhenEditorFails(t *testing.T) {
	m := newActivityTestModel()
	buffer := m.activeBuffer
	m.buffers[buffer].Chat.SetInputValue("original")
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}

	m.finishEditingInput(editorFinishedMsg{buffer: buffer, path: path, err: errors.New("exit status 1")})

	if got := m.buffers[buffer].Chat.InputValue(); got != "original" {
		t.Fatalf("input value = %q, want original", got)
	}
}

func TestEditShortcutRequestsEditor(t *testing.T) {
	for _, controlE := range []bool{false, true} {
		m := newActivityTestModel()

		if msg := m.handleKeybindings(keyPress('x', true)); msg != nil {
			t.Fatalf("ctrl+x returned %T, want nil", msg)
		}
		msg := m.handleKeybindings(keyPress('e', controlE))

		if _, ok := msg.(palette.EditInEditorMsg); !ok {
			t.Fatalf("ctrl+x e (control=%t) returned %T, want palette.EditInEditorMsg", controlE, msg)
		}
	}
}

func TestPaletteEditMessageRequestsEditor(t *testing.T) {
	m := newActivityTestModel()
	t.Setenv("EDITOR", "")

	if cmd := m.editActiveInput(); cmd != nil {
		t.Fatal("editActiveInput returned a command without $EDITOR")
	}
}

func keyPress(code rune, control bool) tea.KeyPressMsg {
	msg := tea.KeyPressMsg{Code: code}
	if control {
		msg.Mod = tea.ModCtrl
	}
	return msg
}
