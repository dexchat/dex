package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/layout"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Model struct {
	input textinput.Model
	theme styles.Theme
}

func New(theme styles.Theme) Model {
	ti := textinput.New()
	ti.Placeholder = "Send message..."
	ti.Focus()
	ti.CharLimit = 256

	return Model{
		input: ti,
		theme: theme,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			m.input.SetValue("")
		}
	}

	return m, cmd
}

func (m Model) View(width int, height int) string {
	inputView := m.input.View()

	if m.input.Value() == "" {
		inputView = m.theme.Styles.InputField.Render(m.input.Placeholder)
	}

	return m.theme.Styles.InputField.
		Border(lipgloss.RoundedBorder()).
		Padding(0, layout.InputBoxPadding).
		Width(width).Height(height).Render(" " + inputView)

}
