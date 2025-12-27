package users

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Model struct {
	Users []string
	theme styles.Theme
}

func New(theme styles.Theme) Model {
	return Model{
		Users: []string{
			"@richard",
			"+gilfoyle",
			"monica",
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
	// Member count header
	memberCount := fmt.Sprintf("%d users", len(m.Users))
	memberHeader := lipgloss.NewStyle().
		Background(m.theme.Colors.Background).
		Foreground(m.theme.Colors.Accent).
		Bold(true).
		Width(width).
		Align(lipgloss.Center).
		Inline(false).
		Render(memberCount)

	dividerLength := width - 2
	if dividerLength < 0 {
		dividerLength = 0
	}
	divider := lipgloss.NewStyle().
		Foreground(m.theme.Colors.LighterBackground).
		Render(strings.Repeat("─", dividerLength))

	var rendered []string

	for _, user := range m.Users {
		rendered = append(rendered, m.theme.Styles.Usernames.Render(user))
	}

	body := memberHeader + "\n" + divider + "\n" + strings.Join(rendered, "\n")

	return m.theme.Styles.Sidebar.
		BorderLeftForeground(m.theme.Colors.LighterBackground).
		Width(width).
		Height(height).
		Render(body)
}
