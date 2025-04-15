package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/channels"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
	"github.com/vaaleyard/dex/internal/ui/components/input"
	"github.com/vaaleyard/dex/internal/ui/components/tabs"
	"github.com/vaaleyard/dex/internal/ui/components/users"
)

type Model struct {
	tabs     tabs.Model
	chat     chat.Model
	users    users.Model
	channels channels.Model
	input    input.Model
}

func New() Model {
	return Model{
		tabs:     tabs.New([]string{"freenode", "libera", "QuakeNet"}),
		chat:     chat.New(),
		users:    users.New(),
		channels: channels.New([]string{"#golang", "#dev", "#linux"}),
		input:    input.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd
	m.tabs, _ = m.tabs.Update(msg)
	m.chat, _ = m.chat.Update(msg)
	m.users, _ = m.users.Update(msg)
	m.channels, _ = m.channels.Update(msg)
	m.input, _ = m.input.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	main := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.channels.View(),
		m.chat.View(),
		m.users.View(),
	)

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		m.tabs.View(),
		main,
		m.input.View(),
		renderFooter(),
	)

	return ui
}

func renderFooter() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5c6370")).
		PaddingTop(1).
		Render("Press q to quit • ↑/↓ to scroll • ←/→ to switch server")
}
