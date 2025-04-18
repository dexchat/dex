package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Model struct {
	Input textinput.Model
}

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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			m.Input.SetValue("")
		}
	}

	return m, cmd
}

func (m Model) View(width int, height int) string {
	return styles.InputBoxStyle.
		Width(width).
		Height(height).
		Render(m.Input.View())
}
