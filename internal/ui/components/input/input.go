package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Input textinput.Model
}

var style = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#3e4452")).
	Background(lipgloss.Color("#1e1e1e")).
	Padding(0, 1).
	MarginTop(1)

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Prompt = "> "
	ti.Width = 80

	return Model{Input: ti}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)

	// Optionally clear input on Enter
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			// In the future: emit IRCMessageSend{m.Input.Value()}
			m.Input.SetValue("") // Clear input
		}
	}

	return m, cmd
}

func (m Model) View() string {
	return style.Render(m.Input.View())
}
