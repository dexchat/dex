package ui

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/ui/components/palette"
)

func TestNormalizeEditorText(t *testing.T) {
	if got := normalizeEditorText("first\r\nsecond\n"); got != "first second" {
		t.Fatalf("normalized text = %q", got)
	}
}

func TestEditorCommandRunsAndCleansUp(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("TMPDIR", dir)
			script := `test "$(cat "$1")" = original || exit 2; printf 'edited\nmessage\n' > "$1"`
			if fail {
				script += "; exit 1"
			}
			command := &editorCommand{Cmd: exec.Command("sh", "-c", script, "test-editor"), draft: "original"}
			err := command.Run()
			if (err != nil) != fail {
				t.Fatalf("Run() error = %v, want failure %t", err, fail)
			}
			if !fail && command.text != "edited message" {
				t.Fatalf("result = %q", command.text)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("temporary files remain: %v, %v", entries, err)
			}
		})
	}
}

func TestFinishEditingInputUsesOriginalBuffer(t *testing.T) {
	m := newActivityTestModel()
	origin := m.getActiveBuffer()
	m.activeBuffer = makeBufferKey("libera", "#go")
	m.Update(editorFinishedMsg{buffer: origin, text: "edited"})
	if origin.Chat.InputValue() != "edited" || m.getActiveBuffer().Chat.InputValue() != "" {
		t.Fatal("result was not applied exclusively to original buffer")
	}
}

func TestFinishEditingInputPreservesDraft(t *testing.T) {
	for _, scenario := range []string{"failure", "changed", "replaced", "removed"} {
		t.Run(scenario, func(t *testing.T) {
			m := newActivityTestModel()
			origin := m.getActiveBuffer()
			origin.Chat.SetInputValue("original")
			msg := editorFinishedMsg{buffer: origin, draft: "original", text: "edited"}
			switch scenario {
			case "failure":
				msg.err = errors.New("editor failed")
			case "changed":
				origin.Chat.SetInputValue("new draft")
			case "replaced":
				replacement := *origin
				m.buffers[origin.Key] = &replacement
			case "removed":
				delete(m.buffers, origin.Key)
			}
			before := origin.Chat.InputValue()
			m.finishEditingInput(msg)
			if origin.Chat.InputValue() != before {
				t.Fatal("draft overwritten")
			}
		})
	}
}

func TestEditShortcutRequestsEditor(t *testing.T) {
	for _, controlE := range []bool{false, true} {
		m := newActivityTestModel()
		if msg := m.handleKeybindings(keyPress('x', true)); msg != nil {
			t.Fatalf("ctrl+x returned %T", msg)
		}
		if msg := m.handleKeybindings(keyPress('e', controlE)); msg != (palette.EditInEditorMsg{}) {
			t.Fatalf("shortcut returned %T", msg)
		}
	}
}

func TestEditRequestDoesNotCreateFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	t.Setenv("EDITOR", "vi")
	m := newActivityTestModel()
	if m.editActiveInput() == nil {
		t.Fatal("missing editor command")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("request performed file I/O: %v, %v", entries, err)
	}
	t.Setenv("EDITOR", "")
	if m.editActiveInput() != nil {
		t.Fatal("command returned without EDITOR")
	}
}

func keyPress(code rune, control bool) tea.KeyPressMsg {
	msg := tea.KeyPressMsg{Code: code}
	if control {
		msg.Mod = tea.ModCtrl
	}
	return msg
}
