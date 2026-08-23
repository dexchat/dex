package ui

import (
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type editorFinishedMsg struct {
	buffer BufferKey
	path   string
	err    error
}

func (m *Model) editActiveInput() tea.Cmd {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		m.addCommandError(m.getActiveBuffer(), "error: $EDITOR is not set")
		return nil
	}

	buffer := m.activeBuffer
	file, err := os.CreateTemp("", "dex-message-*.txt")
	if err != nil {
		m.addCommandError(m.getActiveBuffer(), "error: could not create editor file: "+err.Error())
		return nil
	}
	path := file.Name()
	if _, err = file.WriteString(m.getActiveBuffer().Chat.InputValue()); err == nil {
		err = file.Close()
	} else {
		_ = file.Close()
	}
	if err != nil {
		_ = os.Remove(path)
		m.addCommandError(m.getActiveBuffer(), "error: could not prepare editor file: "+err.Error())
		return nil
	}

	command := exec.Command("sh", "-c", `exec $EDITOR "$1"`, "dex-editor", path)
	command.Env = append(os.Environ(), "EDITOR="+editor)
	return tea.ExecProcess(command, func(err error) tea.Msg {
		return editorFinishedMsg{buffer: buffer, path: path, err: err}
	})
}

func (m *Model) finishEditingInput(msg editorFinishedMsg) {
	defer os.Remove(msg.path)

	buffer, ok := m.buffers[msg.buffer]
	if !ok {
		return
	}
	if msg.err != nil {
		m.addCommandError(buffer, "error: editor failed: "+msg.err.Error())
		return
	}

	contents, err := os.ReadFile(msg.path)
	if err != nil {
		m.addCommandError(buffer, "error: could not read editor file: "+err.Error())
		return
	}
	buffer.Chat.SetInputValue(normalizeEditorText(string(contents)))
}

func normalizeEditorText(value string) string {
	value = strings.TrimRight(value, "\r\n")
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", " ")
}
