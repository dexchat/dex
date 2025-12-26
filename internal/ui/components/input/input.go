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
	ti.CharLimit = 256
	ti.Focus()
	ti.Prompt = ""
	ti.Placeholder = "Send message..."
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Background(theme.Colors.LighterBackground).
		Foreground(theme.Colors.Text)
	ti.TextStyle = lipgloss.NewStyle().
		Background(theme.Colors.LighterBackground).
		Foreground(theme.Colors.Text)

	return Model{
		input: ti,
		theme: theme,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m Model) View(width int) string {
	inputWidth := width - 1
	if inputWidth < 12 { // random number to not panic the app
		inputWidth = 12
	}
	m.input.Width = inputWidth

	inputView := m.input.View()
	inputContainer := m.theme.Styles.InputField.
		Margin(0, layout.InputBoxMargin, 0, layout.InputBoxMargin).
		MarginBackground(m.theme.Colors.Background).
		Border(lipgloss.NormalBorder()).
		BorderBackground(m.theme.Colors.LighterBackground).
		BorderForeground(m.theme.Colors.LighterBackground).
		Width(width).Render(inputView)

	return inputContainer
}
