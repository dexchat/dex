package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	Servers []string
	Focused int
}

var (
	tabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Underline(true)
	tabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	tabGap      = "   "
)

func New(servers []string) Model {
	return Model{Servers: servers, Focused: 0}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.Focused > 0 {
				m.Focused--
			}
		case "right":
			if m.Focused < len(m.Servers)-1 {
				m.Focused++
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var parts []string
	for i, t := range m.Servers {
		if i == m.Focused {
			parts = append(parts, tabActive.Render(t))
		} else {
			parts = append(parts, tabInactive.Render(t))
		}
	}
	return strings.Join(parts, tabGap)
}
