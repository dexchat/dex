package users

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Users []string
}

var style = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 2).
	Width(20).
	BorderForeground(lipgloss.Color("63"))

func New() Model {
	return Model{
		Users: []string{
			"@leo",
			"+bob",
			"alice",
			"charlie",
			"dave",
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// No interaction for now
	return m, nil
}

func (m Model) View() string {
	content := strings.Join(m.Users, "\n")
	return style.Render(content)
}
