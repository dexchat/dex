package chat

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dexchat/dex/internal/ui/components/keybindings"
	"github.com/dexchat/dex/internal/ui/styles"
)

// SendAction is a synchronous request handled by the owning UI model.
type SendAction struct {
	Text string
}

type Model struct {
	viewport viewport.Model
	topic    string
	messages []Message
	nickname string
	input    textinput.Model

	// Rendered content is cached so queued messages only render the new tail of
	// the chat history. Changes that affect existing lines invalidate this cache.
	needsRender      bool
	renderedLines    []string
	renderedMessages int
	renderWidth      int

	channelMembers ChannelMembers
	completion     nicknameCompletion

	usernameColors styles.UsernameColors
	theme          styles.Theme
}

func New(theme styles.Theme, usernameColors styles.UsernameColors) Model {
	input := textinput.New()
	input.CharLimit = 256
	input.Focus()
	input.Prompt = ""
	input.Placeholder = "Send message..."
	inputStyles := input.Styles()
	inputStyles.Focused.Placeholder = theme.Styles.InputField
	inputStyles.Focused.Text = theme.Styles.InputField
	inputStyles.Blurred.Placeholder = theme.Styles.InputField
	inputStyles.Blurred.Text = theme.Styles.InputField
	input.SetStyles(inputStyles)

	m := Model{
		messages:       make([]Message, 0),
		topic:          "",
		theme:          theme,
		usernameColors: usernameColors,
		input:          input,
		viewport:       viewport.New(viewport.WithWidth(0), viewport.WithHeight(0)),
		nickname:       "",
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) InputValue() string {
	return m.input.Value()
}

func (m *Model) SetInputValue(value string) {
	m.input.SetValue(value)
	m.input.CursorEnd()
}

func (m *Model) Update(msg tea.Msg) (Model, *SendAction, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.input.Focused() {
			keyMap := keybindings.DefaultKeyMap()
			if key.Matches(msg, keyMap.Autocomplete) {
				m.completeNickname()
				return *m, nil, nil
			}
			m.completion = nicknameCompletion{}

			if msg.Code == tea.KeyEnter && m.input.Value() != "" {
				text := m.input.Value()
				m.input.Reset()
				return *m, &SendAction{Text: text}, nil
			}

			if key.Matches(msg, keyMap.ScrollUp, keyMap.ScrollDown) {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
				return *m, nil, tea.Batch(cmds...)
			}

			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
			return *m, nil, tea.Batch(cmds...)
		}
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return *m, nil, tea.Batch(cmds...)
}

func (m *Model) View() string {
	topicView := m.renderTopic(m.viewport.Width())
	chatInputBox := m.renderInputBox()

	// The viewport content is set by updateContent
	chatViewport := lipgloss.NewStyle().
		Background(m.theme.Colors.Base.Background).
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
	chatAreaFrameHeight := m.theme.Styles.ChatArea.GetVerticalFrameSize()
	viewportHeight := height - topicHeight - inputHeight - chatAreaFrameHeight

	if viewportHeight < 3 {
		viewportHeight = 3 // Minimum height for viewport
	}

	widthChanged := m.viewport.Width() != width
	heightChanged := m.viewport.Height() != viewportHeight
	if !widthChanged && !heightChanged && !m.needsRender {
		return
	}

	if widthChanged {
		m.viewport.SetWidth(width)
		m.invalidateRenderedContent()
	}
	if heightChanged {
		m.viewport.SetHeight(viewportHeight)
	}
	if m.needsRender {
		m.updateContent()
	}
}

func (m *Model) updateContent() {
	width := m.viewport.Width()
	if m.renderWidth != width || m.renderedMessages > len(m.messages) {
		m.renderedLines = nil
		m.renderedMessages = 0
		m.renderWidth = width
	}

	var lastTimestamp time.Time
	if m.renderedMessages > 0 {
		lastTimestamp = m.messages[m.renderedMessages-1].Timestamp
	}

	for _, msg := range m.messages[m.renderedMessages:] {
		if !lastTimestamp.IsZero() && !sameDay(lastTimestamp, msg.Timestamp) {
			m.renderedLines = append(m.renderedLines, strings.Split(m.renderDateSeparator(msg.Timestamp, width), "\n")...)
		}
		lastTimestamp = msg.Timestamp
		m.renderedLines = append(m.renderedLines, strings.Split(m.renderMessage(msg, width), "\n")...)
	}

	wasAtBottom := m.viewport.AtBottom()
	m.viewport.SetContentLines(m.renderedLines)
	if wasAtBottom {
		m.viewport.GotoBottom()
	}
	m.renderedMessages = len(m.messages)
	m.renderWidth = width
	m.needsRender = false
}

func (m *Model) invalidateRenderedContent() {
	m.renderedLines = nil
	m.renderedMessages = 0
	m.renderWidth = 0
	m.needsRender = true
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func (m *Model) renderDateSeparator(date time.Time, width int) string {
	dateStr := date.Format("Mon, January 2")
	padding := (width - len(dateStr) - 2) / 2
	if padding < 1 {
		padding = 1
	}
	dashes := strings.Repeat("─", padding)
	separator := dashes + " " + dateStr + " " + dashes

	style := lipgloss.NewStyle().
		Background(m.theme.Colors.Base.Background).
		Foreground(m.theme.Colors.Base.Dimmed)

	return style.Width(width).Render(separator)
}

func (m *Model) AddMessage(msg Message) {
	m.QueueMessage(msg)
	m.FlushQueue()
}

func (m *Model) QueueMessage(msg Message) {
	m.messages = append(m.messages, msg)
	m.needsRender = true
}

// ReplaceMessages swaps the complete chat history without rendering it on the
// current update. History loading uses this to avoid blocking the UI.
func (m *Model) ReplaceMessages(messages []Message) {
	m.messages = messages
	m.invalidateRenderedContent()
}

func (m *Model) FlushQueue() {
	if m.needsRender {
		m.updateContent()
	}
}

func (m *Model) SetTopic(topic string) {
	m.topic = topic
}

func (m *Model) SetNickname(nickname string) {
	if m.nickname == nickname {
		return
	}
	m.nickname = nickname
	m.setInputWidth(m.viewport.Width())
	m.invalidateRenderedContent()
}

func (m *Model) SetTheme(theme styles.Theme, usernameColors styles.UsernameColors) {
	m.theme = theme
	m.usernameColors = usernameColors
	inputStyles := m.input.Styles()
	inputStyles.Focused.Placeholder = theme.Styles.InputField
	inputStyles.Focused.Text = theme.Styles.InputField
	inputStyles.Blurred.Placeholder = theme.Styles.InputField
	inputStyles.Blurred.Text = theme.Styles.InputField
	m.input.SetStyles(inputStyles)
	m.invalidateRenderedContent()
	if m.viewport.Width() > 0 {
		m.updateContent()
	}
}

func (m *Model) Nickname() string {
	return m.nickname
}

func (m *Model) RefreshContent() {
	m.invalidateRenderedContent()
	m.FlushQueue()
}

func (m *Model) SetChannelMembers(members ChannelMembers) {
	m.channelMembers = members
}
