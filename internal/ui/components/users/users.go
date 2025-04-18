package users

import (
	"github.com/vaaleyard/dex/internal/ui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Users []string
	theme styles.Theme
}

func New(theme styles.Theme) Model {
	return Model{
		Users: []string{
			"@leo",
			"+amora",
			"gilfoyle",
			"dinesh",
			"jared",
		},
		theme: theme,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// No interaction for now
	return m, nil
}

func (m Model) View(width int, height int) string {
	var rendered []string

	for _, user := range m.Users {
		rendered = append(rendered, styles.UsernameStyle(user))
	}

	body := strings.Join(rendered, "\n")

	return m.theme.Styles.Sidebar.
		Width(width).
		Height(height).
		Render(body)
}
