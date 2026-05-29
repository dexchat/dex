package users

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))
	vp.Style = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background)

	return Model{
		users:          make([]string, 0),
		viewport:       vp,
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
	rendered := make([]string, len(m.users))
	baseStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Base.Background).
		Width(m.viewport.Width())

	for i, user := range m.users {
		nick := strings.TrimLeft(user, "~&@%+")
		rendered[i] = baseStyle.
			Foreground(m.usernameColors.GetColor(nick)).
			Render(user)
	}
	m.viewport.SetContent(strings.Join(rendered, "\n"))

	return m
}

func (m Model) SetSize(width, height int) Model {
	contentWidth := width - usersListVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	viewportHeight := height - m.theme.Styles.Sidebar.GetVerticalFrameSize() - headerHeight
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(viewportHeight)
	return m.updateContent()
}

func (m Model) HasUser(nick string) bool {
	for _, u := range m.users {
		stripped := strings.TrimLeft(u, "~&@%+")
		if strings.EqualFold(stripped, nick) {
			return true
		}
	}
	return false
}

func (m Model) GetUserPrefix(nick string) string {
	for _, u := range m.users {
		stripped := strings.TrimLeft(u, "~&@%+")
		if strings.EqualFold(stripped, nick) {
			return u[:len(u)-len(stripped)]
		}
	}
	return ""
}

func (m Model) View(width, height int) string {
	contentWidth := width - usersListVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	viewportHeight := height - m.theme.Styles.Sidebar.GetVerticalFrameSize() - headerHeight
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	if m.viewport.Width() != contentWidth || m.viewport.Height() != viewportHeight {
		m = m.SetSize(width, height)
	}

	var body string
	// Show member list/count only on channels
	if len(m.users) > 0 {
		memberCount := fmt.Sprintf("%d users", len(m.users))
		memberHeader := lipgloss.NewStyle().
			Background(m.theme.Colors.Base.Background).
			Foreground(m.theme.Colors.Base.Accent).
			Bold(true).
			Width(contentWidth).
			Align(lipgloss.Center).
			Inline(false).
			Render(memberCount)

		divider := lipgloss.NewStyle().
			Foreground(m.theme.Colors.Base.Surface).
			Render(strings.Repeat("─", contentWidth))

		body = memberHeader + "\n" + divider + "\n" + m.viewport.View()
	} else {
		body = m.viewport.View()
	}

	return m.theme.Styles.Sidebar.
		Height(height).
		BorderLeftForeground(m.theme.Colors.Base.Surface).
		Render(body)
}
