package ui

import (
	"github.com/vaaleyard/dex/internal/ui/layout"
	"github.com/vaaleyard/dex/internal/ui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/channels"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
	"github.com/vaaleyard/dex/internal/ui/components/users"
)

const (
	// Arbitrary values that I think looks good in the screen
	// I believe that having a slightly bigger chat panel compared to
	// the user panel is better for the eyes
	channelsPanelMaxWidth = 25
	usersPanelMaxWidth    = 20
	// 2 lines of padding to not overflow the texts to the top
	topPadding = 2
	// channels right (1) + users left (1) + chat borders (2)
	appVerticalBordersSize = 4
)

type Model struct {
	// TODO: replace for Width and Height only
	layout         layout.Layout
	theme          styles.Theme
	usernameColors styles.UsernameColors

	chat     chat.Model
	users    users.Model
	channels channels.Model
}

func New() *Model {
	m := Model{}

	m.theme = styles.AyuDarkTheme()
	m.usernameColors = styles.NewUsernameColors(m.theme.Colors.Usernames)

	m.channels = channels.New(m.theme)
	m.chat = chat.New(m.theme, m.usernameColors)
	m.users = users.New(m.theme, m.usernameColors)

	return &m
}

func (m *Model) Init() tea.Cmd {
	return m.chat.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.layout.Width = msg.Width
		m.layout.Height = msg.Height
		adjustedHeight := m.layout.Height - topPadding
		if adjustedHeight < 0 {
			adjustedHeight = 0
		}

		chatWidth := m.layout.Width - channelsPanelMaxWidth - usersPanelMaxWidth - appVerticalBordersSize
		m.chat.SetSize(chatWidth, adjustedHeight)
	}

	m.chat, cmd = m.chat.Update(msg)
	cmds = append(cmds, cmd)

	m.users, cmd = m.users.Update(msg)
	cmds = append(cmds, cmd)

	m.channels, cmd = m.channels.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	adjustedHeight := m.layout.Height - topPadding
	if adjustedHeight < 0 {
		adjustedHeight = 0
	}

	channelsView := lipgloss.NewStyle().
		MaxWidth(channelsPanelMaxWidth).
		PaddingTop(topPadding).
		Render(m.channels.View(channelsPanelMaxWidth, adjustedHeight))

	usersView := lipgloss.NewStyle().
		MaxWidth(usersPanelMaxWidth).
		PaddingTop(topPadding).
		Render(m.users.View(usersPanelMaxWidth, adjustedHeight))

	chatView := lipgloss.NewStyle().
		PaddingTop(topPadding).
		Height(adjustedHeight).
		Render(m.chat.View())

	return lipgloss.JoinHorizontal(lipgloss.Top,
		channelsView,
		chatView,
		usersView,
	)
}
