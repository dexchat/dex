package users

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	// left (1) + right (1) borders
	usersListVerticalBordersSize = 2
	// header + divider
	headerHeight = 2
)

// UserListMsg is a message received from the IRC client containing the updated user list to be displayed
type UserListMsg []string

type Model struct {
	users          []string
	viewport       viewport.Model
	theme          styles.Theme
	usernameColors styles.UsernameColors
}

func New(theme styles.Theme, usernameColors styles.UsernameColors) Model {
	return Model{
		users:          make([]string, 0),
		viewport:       viewport.New(0, 0),
		theme:          theme,
		usernameColors: usernameColors,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case UserListMsg:
		m.users = msg
		return m.updateContent(), nil
	}
	// This prevents j/k in the input box from scrolling the viewport
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, cmd
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) updateContent() Model {
	var rendered []string
	for _, user := range m.users {
		userStyle := lipgloss.NewStyle().
			Background(m.theme.Colors.Background).
			Width(m.viewport.Width).
			Foreground(m.usernameColors.GetColor(user))
		rendered = append(rendered, userStyle.Render(user))
	}
	m.viewport.SetContent(strings.Join(rendered, "\n"))

	return m
}

func (m Model) SetSize(width, height int) Model {
	contentWidth := width - usersListVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	viewportHeight := height - headerHeight
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	m.viewport.Width = contentWidth
	m.viewport.Height = viewportHeight
	return m.updateContent()
}

func (m Model) View(width, height int) string {
	contentWidth := width - usersListVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	memberCount := fmt.Sprintf("%d users", len(m.users))
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

	body := memberHeader + "\n" + divider + "\n" + m.viewport.View()

	return m.theme.Styles.Sidebar.
		Height(height).
		BorderLeftForeground(m.theme.Colors.LighterBackground).
		Render(body)
}
