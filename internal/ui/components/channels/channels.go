package channels

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Channels     []string
	SelectedChan int
}

var style = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 2).
	Width(20).
	BorderForeground(lipgloss.Color("36"))

var selectedStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("205"))

func New(channels []string) Model {
	return Model{
		Channels:     channels,
		SelectedChan: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.SelectedChan > 0 {
				m.SelectedChan--
			}
		case "down":
			if m.SelectedChan < len(m.Channels)-1 {
				m.SelectedChan++
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var rendered []string
	for i, ch := range m.Channels {
		if i == m.SelectedChan {
			rendered = append(rendered, selectedStyle.Render("→ "+ch))
		} else {
			rendered = append(rendered, "  "+ch)
		}
	}
	return style.Render(strings.Join(rendered, "\n"))
}
