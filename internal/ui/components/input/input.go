package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/layout"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Model struct {
	input    textinput.Model
	theme    styles.Theme
	nickname string
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
		input:    ti,
		theme:    theme,
		nickname: "trump",
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
	nicknamePrefix := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Accent).
		Bold(true).
		Render(m.nickname)

	prompt := lipgloss.NewStyle().
		Foreground(m.theme.Colors.BorderColor).
		Background(m.theme.Colors.LighterBackground).
		Render(" | ")

	prefixWidth := len(m.nickname) + len(" > ")
	inputWidth := width - prefixWidth - layout.InputBoxMargin*2 - 3 // borders and margins

	if inputWidth < 12 { // minimum width to not panic
		inputWidth = 12
	}
	m.input.Width = inputWidth

	inputView := nicknamePrefix + prompt + m.input.View()

	inputContainer := m.theme.Styles.InputField.
		Margin(0, layout.InputBoxMargin, 0, layout.InputBoxMargin).
		MarginBackground(m.theme.Colors.Background).
		Border(lipgloss.NormalBorder()).
		BorderBackground(m.theme.Colors.LighterBackground).
		BorderForeground(m.theme.Colors.LighterBackground).
		Width(width).Render(inputView)

	return inputContainer
}
