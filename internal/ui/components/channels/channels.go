package channels

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Channels []string
	Selected int

	theme styles.Theme
}

func New(theme styles.Theme, channels []string) Model {
	return Model{
		Channels: channels,
		Selected: 0,
		theme:    theme,
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
			if m.Selected > 0 {
				m.Selected--
			}
		case "down":
			if m.Selected < len(m.Channels)-1 {
				m.Selected++
			}
		}
	}
	return m, nil
}

func (m Model) View(width int, height int) string {
	var rendered []string
	for i, ch := range m.Channels {
		if i == m.Selected {
			rendered = append(rendered, lipgloss.NewStyle().Render("→ "+ch))
		} else {
			rendered = append(rendered, "  "+ch)
		}
	}
	body := strings.Join(rendered, "\n")

	return m.theme.Styles.Sidebar.
		Width(width).
		Height(height).
		Render(body)
}
