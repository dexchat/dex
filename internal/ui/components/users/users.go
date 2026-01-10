package users

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	// left (1) + right (1) borders
	usersListVerticalBordersSize = 2
)

// UserListMsg is a message received from the IRC client containing the new user list.
type UserListMsg []string

type Model struct {
	Users          []string
	theme          styles.Theme
	usernameColors styles.UsernameColors
}

func New(theme styles.Theme, usernameColors styles.UsernameColors) Model {
	return Model{
		Users:          make([]string, 0),
		theme:          theme,
		usernameColors: usernameColors,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UserListMsg:
		m.Users = msg
		return m, nil
	}
	return m, nil
}

func (m Model) View(width, height int) string {
	contentWidth := width - usersListVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	// Member count header
	memberCount := fmt.Sprintf("%d users", len(m.Users))
	memberHeader := lipgloss.NewStyle().
		Background(m.theme.Colors.Background).
		Foreground(m.theme.Colors.Accent).
		Bold(true).
		Width(contentWidth).
		Align(lipgloss.Center).
		Inline(false).
		Render(memberCount)

	divider := lipgloss.NewStyle().
		Foreground(m.theme.Colors.LighterBackground).
		Render(strings.Repeat("─", contentWidth))

	var rendered []string

	for _, user := range m.Users {
		userStyle := lipgloss.NewStyle().
			Foreground(m.usernameColors.GetColor(user))
		rendered = append(rendered, userStyle.Render(user))
	}

	body := memberHeader + "\n" + divider + "\n" + strings.Join(rendered, "\n")

	return m.theme.Styles.Sidebar.
		Height(height).
		BorderLeftForeground(m.theme.Colors.LighterBackground).
		Render(body)
}
