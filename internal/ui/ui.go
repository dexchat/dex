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
	"github.com/vaaleyard/dex/internal/ui/components/input"
	"github.com/vaaleyard/dex/internal/ui/components/servers"
	"github.com/vaaleyard/dex/internal/ui/components/users"
)

type Model struct {
	layout layout.Layout
	theme  styles.Theme

	servers  servers.Model
	chat     chat.Model
	users    users.Model
	channels channels.Model
	input    input.Model
}

func New() Model {
	m := Model{}

	m.theme = styles.AyuDarkTheme()
	m.servers = servers.New([]string{"freenode", "libera", "QuakeNet"})
	m.channels = channels.New([]string{"#golang", "#dev", "#linux", "#cinema", "#rust", "#docker", "#kubernetes", "#alpine"})
	m.chat = chat.New()
	m.users = users.New()
	m.input = input.New(m.theme)

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.layout = layout.GenerateLayout(msg.Width, msg.Height, m.servers.Hidden)

		s := fmt.Sprintf("Terminal size: %dx%d\n"+
			"Channel Width: %d\n"+
			"Chat width: %d\n"+
			"User width: %d\n"+
			"Middle content sum: %d\n"+
			"input width: %d\n", msg.Width, msg.Height, m.layout.ChannelSidebarWidth, m.layout.MainContentWidth, m.layout.UserSidebarWidth,
			m.layout.ChannelSidebarWidth+m.layout.MainContentWidth+m.layout.UserSidebarWidth, m.layout.ScreenWidth)

		f, _ := os.Create("debug.log")
		_, _ = f.WriteString(s)
	}

	var cmds []tea.Cmd
	m.servers, _ = m.servers.Update(msg)
	m.chat, _ = m.chat.Update(msg)
	m.users, _ = m.users.Update(msg)
	m.channels, _ = m.channels.Update(msg)
	m.input, _ = m.input.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	middleContent := lipgloss.JoinHorizontal(lipgloss.Top,
		m.channels.View(m.layout.ChannelSidebarWidth, m.layout.MainContentHeight),
		m.chat.View(m.layout.MainContentWidth, m.layout.MainContentHeight),
		m.users.View(m.layout.UserSidebarWidth, m.layout.MainContentHeight),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		m.servers.View(m.layout.ScreenWidth, m.layout.TabRowHeight),
		middleContent,
		m.input.View(m.layout.InputBoxWidth, m.layout.InputBoxHeight),
	)
}
