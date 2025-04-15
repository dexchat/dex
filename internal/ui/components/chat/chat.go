package chat

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Messages []string
}

var style = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#3e4452")).
	Padding(1, 2).
	Margin(0, 1)

func New() Model {
	return Model{
		Messages: []string{
			"[10:00] <alice> hello",
			"[10:01] <bob> hi",
			"[10:02] <you> what's up?",
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// for now: no interactivity
	return m, nil
}

func (m Model) View() string {
	return style.Render(strings.Join(m.Messages, "\n"))
}
