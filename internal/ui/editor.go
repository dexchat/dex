package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type editorFinishedMsg struct {
	buffer *Buffer
	draft  string
	text   string
	err    error
}

// editorCommand owns the file and process while Bubble Tea releases the terminal.
type editorCommand struct {
	*exec.Cmd
	draft string
	text  string
}

func (c *editorCommand) SetStdin(r io.Reader)  { c.Stdin = r }
func (c *editorCommand) SetStdout(w io.Writer) { c.Stdout = w }
func (c *editorCommand) SetStderr(w io.Writer) { c.Stderr = w }

func (c *editorCommand) Run() (err error) {
	file, err := os.CreateTemp("", "dex-message-*.txt")
	if err != nil {
		return fmt.Errorf("could not create editor file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("could not remove editor file %s: %w", file.Name(), removeErr))
		}
	}()
	_, writeErr := file.WriteString(c.draft)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return fmt.Errorf("could not prepare editor file: %w", err)
	}
	c.Args = append(c.Args, file.Name())
	if err := c.Cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	contents, err := os.ReadFile(file.Name())
	if err != nil {
		return fmt.Errorf("could not read editor file: %w", err)
	}
	c.text = normalizeEditorText(string(contents))
	return nil
}

func (m *Model) editActiveInput() tea.Cmd {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		m.addCommandError(m.getActiveBuffer(), "error: $EDITOR is not set")
		return nil
	}
	buffer := m.getActiveBuffer()
	draft := buffer.Chat.InputValue()
	command := &editorCommand{Cmd: exec.Command("sh", "-c", `exec $EDITOR "$1"`, "dex-editor"), draft: draft}
	command.Env = append(os.Environ(), "EDITOR="+editor)
	return tea.Exec(command, func(err error) tea.Msg {
		return editorFinishedMsg{buffer: buffer, draft: draft, text: command.text, err: err}
	})
}

func (m *Model) finishEditingInput(msg editorFinishedMsg) {
	if m.buffers[msg.buffer.Key] != msg.buffer {
		return
	}
	if msg.err != nil {
		m.addCommandError(msg.buffer, "error: "+msg.err.Error())
	} else if msg.buffer.Chat.InputValue() != msg.draft {
		m.addCommandError(msg.buffer, "error: draft changed while editing; current draft was preserved")
	} else {
		msg.buffer.Chat.SetInputValue(msg.text)
	}
}

func normalizeEditorText(value string) string {
	value = strings.TrimRight(value, "\r\n")
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", " ")
}
