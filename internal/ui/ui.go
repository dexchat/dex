package ui

import (
	"github.com/vaaleyard/dex/internal/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	overlay "github.com/rmhubbert/bubbletea-overlay"
	"github.com/vaaleyard/dex/internal/ui/components/channels"
	"github.com/vaaleyard/dex/internal/ui/components/chat"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/components/palette"
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

type model struct {
	Height         int
	theme          styles.Theme
	usernameColors styles.UsernameColors

	chat     chat.Model
	users    users.Model
	channels channels.Model
	palette  *palette.Model
	overlay  *overlay.Model
}

func New() *model {
	m := model{}

	m.theme = styles.AyuDarkTheme()
	m.usernameColors = styles.NewUsernameColors(m.theme.Colors.Usernames)

	m.channels = channels.New(m.theme)
	m.chat = chat.New(m.theme, m.usernameColors)
	m.users = users.New(m.theme, m.usernameColors)
	m.palette = palette.New(m.theme)

	return &m
}

func (m *model) Init() tea.Cmd {
	return m.chat.Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd        tea.Cmd
		cmds       []tea.Cmd
		paletteMsg tea.Msg = msg
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

		kb := keybindings.DefaultKeyMap()
		switch {
		case key.Matches(msg, kb.MoveDown):
			if !m.palette.IsVisible() {
				m.channels.MoveDown()
			}
		case key.Matches(msg, kb.MoveUp):
			if !m.palette.IsVisible() {
				m.channels.MoveUp()
			}
		case key.Matches(msg, kb.TogglePalette):
			m.palette.Toggle()
			// Consume the Ctrl+O key so it doesn't immediately close the palette
			paletteMsg = nil
		}
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		adjustedHeight := m.Height - topPadding
		if adjustedHeight < 0 {
			adjustedHeight = 0
		}

		m.palette.SetSize(90, len(m.palette.Commands())+5)

		chatWidth := msg.Width - channelsPanelMaxWidth - usersPanelMaxWidth - appVerticalBordersSize
		m.chat.SetSize(chatWidth, adjustedHeight)
		m.chat.SetContent()
	}

	palModel, cmd := m.palette.Update(paletteMsg)
	m.palette = palModel.(*palette.Model)
	cmds = append(cmds, cmd)

	// do not update background if palette is visible,
	// if it gets updated, the writing will be shared with chat input
	// TODO: update only the viewport
	if !m.palette.IsVisible() {
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)

		m.users, cmd = m.users.Update(msg)
		cmds = append(cmds, cmd)

		m.channels, cmd = m.channels.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update overlay reference if palette is visible
	if m.palette.IsVisible() {
		m.overlay = overlay.New(m.palette, &background{m}, overlay.Center, overlay.Center, 0, 0)
	} else {
		m.overlay = nil
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.palette.IsVisible() && m.overlay != nil {
		return m.overlay.View()
	}

	return (&background{m}).View()
}
