package ui

import (
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/ui/styles"

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
	config *config.Config

	Height         int
	theme          styles.Theme
	usernameColors styles.UsernameColors

	chat     chat.Model
	users    users.Model
	channels channels.Model
	palette  *palette.Model
	overlay  *overlay.Model
}

func New(cfg *config.Config) *model {
	m := model{}
	m.config = cfg

	m.theme = styles.AyuDarkTheme()
	m.usernameColors = styles.NewUsernameColors(m.theme.Colors.Usernames)

	m.channels = channels.New(m.theme, m.config.Servers)
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
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msgTyped := msg.(type) {
	case tea.KeyMsg:
		if msgTyped.Type == tea.KeyCtrlC {
			if keybindings.QuitHandler() {
				return m, tea.Quit
			}
			return m, nil
		}
		msg = m.handleKeybindings(msgTyped)

	case tea.WindowSizeMsg:
		m.Height = msgTyped.Height
		adjustedHeight := m.Height - topPadding
		if adjustedHeight < 0 {
			adjustedHeight = 0
		}

		// TODO: ideally palette width should be smaller than chat width. It might happen if the font size is too big
		m.palette.SetSize(90, len(m.palette.Commands())+5)

		chatWidth := msgTyped.Width - channelsPanelMaxWidth - usersPanelMaxWidth - appVerticalBordersSize
		m.chat.SetSize(chatWidth, adjustedHeight)
		m.chat.SetContent()
	}

	// Write characters in the palette input bar if it's open, instead of chat input
	if m.palette.IsVisible() {
		palModel, cmd := m.palette.Update(msg)
		m.palette = palModel.(*palette.Model)
		cmds = append(cmds, cmd)
	}

	// Update background components, blocking keyboard input when the palette is open
	msg = m.filterOutKeyMsgs(msg)
	m.chat, cmd = m.chat.Update(msg)
	cmds = append(cmds, cmd)

	m.users, cmd = m.users.Update(msg)
	cmds = append(cmds, cmd)

	m.channels, cmd = m.channels.Update(msg)
	cmds = append(cmds, cmd)

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

// filterOutKeyMsgs filters out keyboard messages from background components when palette is open
func (m *model) filterOutKeyMsgs(msg tea.Msg) tea.Msg {
	// If palette is not visible, pass all messages through
	if !m.palette.IsVisible() {
		return msg
	}

	// If palette is visible, block keyboard input to background components
	if _, isKeyMsg := msg.(tea.KeyMsg); isKeyMsg {
		return nil
	}

	// Allow non-keyboard messages (e.g., WindowSizeMsg) to reach background components
	return msg
}
