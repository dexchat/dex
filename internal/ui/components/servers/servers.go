package servers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

// TODO: is tabs a good feature? maybe it can be put on the channel sidebar

type Model struct {
	style   styles.Tab
	Hidden  bool
	Servers []string
	Focused int
}

func New(servers []string) Model {
	return Model{
		Hidden:  true,
		style:   styles.NewTabStyle(),
		Servers: servers,
		Focused: 0,
	}
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

func (m Model) View(width int, height int) string {
	if m.Hidden {
		return ""
	}
	var parts []string

	for i, name := range m.Servers {
		if i == m.Focused {
			parts = append(parts, m.style.ActiveTab.Render(name))
		} else {
			parts = append(parts, m.style.InactiveTab.Render(name))
		}
	}

	tabLine := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	return m.style.TabBar.Width(width).Height(height).Render(tabLine)
}
