package chat

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Model struct {
	viewport viewport.Model
	topic    string
	messages []Message
	nickname string
	input    textinput.Model

	usernameColors styles.UsernameColors
	theme          styles.Theme
}

func New(theme styles.Theme, usernameColors styles.UsernameColors) Model {
	input := textinput.New()
	input.CharLimit = 256
	input.Focus()
	input.Prompt = ""
	input.Placeholder = "Send message..."
	input.PlaceholderStyle = lipgloss.NewStyle().
		Background(theme.Colors.LighterBackground).
		Foreground(theme.Colors.Text)
	input.TextStyle = lipgloss.NewStyle().
		Background(theme.Colors.LighterBackground).
		Foreground(theme.Colors.Text)

	m := Model{
		messages:       make([]Message, 0),
		topic:          "",
		theme:          theme,
		usernameColors: usernameColors,
		input:          input,
		viewport:       viewport.New(0, 0),
		nickname:       "johnbogle",
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Filters out mouse events from writing in input bar and disable viewport keybindings (j/k)
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if m.input.Focused() &&
			msg.Action == tea.MouseActionPress &&
			(msg.Button == tea.MouseButtonWheelUp ||
				msg.Button == tea.MouseButtonWheelDown) {

			return *m, tea.Batch(cmds...)
		}
	case tea.KeyMsg:
		if m.input.Focused() {
			return *m, tea.Batch(cmds...)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return *m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	topicView := m.renderTopic(m.viewport.Width)
	chatInputBox := m.renderInputBox()

	// The viewport content is set by updateContent
	chatViewport := lipgloss.NewStyle().
		Background(m.theme.Colors.Background).
		PaddingLeft(1).
		PaddingRight(1).
		Render(m.viewport.View())

	combinedView := lipgloss.JoinVertical(
		lipgloss.Left,
		topicView,
		chatViewport,
		chatInputBox,
	)

	return m.theme.Styles.ChatArea.
		Render(combinedView)
}

func (m *Model) SetSize(width, height int) {
	m.setInputWidth(width)

	topicHeight := lipgloss.Height(m.renderTopic(width))
	inputHeight := lipgloss.Height(m.renderInputBox())
	viewportHeight := height - topicHeight - inputHeight

	if viewportHeight < 3 {
		viewportHeight = 3 // Minimum height for viewport
	}

	m.viewport.Width = width
	m.viewport.Height = viewportHeight
	m.updateContent()
}

func (m *Model) updateContent() {
	// Re-render messages with the new width and update viewport content
	styledMessages := make([]string, len(m.messages))
	for i, msg := range m.messages {
		styledMessages[i] = m.renderMessage(msg, m.viewport.Width)
	}
	m.viewport.SetContent(strings.Join(styledMessages, "\n"))
	m.viewport.GotoBottom()
}

func (m *Model) AddMessage(msg Message) {
	m.messages = append(m.messages, msg)
	m.updateContent()
}

func (m *Model) SetTopic(topic string) {
	m.topic = topic
}
