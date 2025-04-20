package ui

import (
	"fmt"
	"github.com/vaaleyard/dex/internal/ui/layout"
	"github.com/vaaleyard/dex/internal/ui/styles"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/channels"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
	"github.com/vaaleyard/dex/internal/ui/components/users"
)

type Model struct {
	layout layout.Layout
	theme  styles.Theme

	chat     chat.Model
	users    users.Model
	channels channels.Model
}

func New() Model {
	m := Model{}

	m.theme = styles.AyuDarkTheme()
	m.channels = channels.New(m.theme, []string{"#golang", "#dev", "#linux", "#cinema", "#rust", "#docker", "#kubernetes", "#alpine"})
	m.chat = chat.New(m.theme)
	m.users = users.New(m.theme)

	return m
}

func (m Model) Init() tea.Cmd {
	return m.chat.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.layout = layout.GenerateLayout(msg.Width, msg.Height)

		s := fmt.Sprintf("1\nTerminal size: %dx%d\n"+
			"Sidebar size: %dx%d\n"+
			"Chat size: %dx%d\n",
			msg.Width, msg.Height,
			m.layout.SiderbarWidth, m.layout.AppHeight,
			m.layout.ChatWidth, m.layout.AppHeight)

		f, _ := os.Create("debug.log")
		_, _ = f.WriteString(s)
	}

	var cmds []tea.Cmd
	m.chat, _ = m.chat.Update(msg)
	m.users, _ = m.users.Update(msg)
	m.channels, _ = m.channels.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.channels.View(m.layout.SiderbarWidth, m.layout.AppHeight),
		m.chat.View(m.layout.ChatWidth, m.layout.AppHeight),
		m.users.View(m.layout.SiderbarWidth, m.layout.AppHeight),
	)
}
