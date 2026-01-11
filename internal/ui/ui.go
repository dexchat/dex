package ui

import (
	"log"

	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/irc"
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

type Model struct {
	config *config.Config

	width          int
	height         int
	theme          styles.Theme
	usernameColors styles.UsernameColors

	buffers      map[BufferKey]*Buffer
	activeBuffer BufferKey

	channels channels.Model
	palette  *palette.Model
	overlay  *overlay.Model
}

func New(cfg *config.Config) *Model {
	m := Model{
		config:  cfg,
		buffers: make(map[BufferKey]*Buffer),
		theme:   styles.AyuDarkTheme(),
	}
	m.usernameColors = styles.NewUsernameColors(m.theme.Colors.Usernames)

	m.channels = channels.New(m.theme, m.config.Servers)
	m.palette = palette.New(m.theme)

	for _, server := range cfg.Servers {
		serverKey := makeBufferKey(server.Name, "")
		m.buffers[serverKey] = &Buffer{
			Key:    serverKey,
			Server: server.Name,
			Chat:   chat.New(m.theme, m.usernameColors),
			Users:  users.New(m.theme, m.usernameColors),
		}

		// Set the nickname configured in the config file before connecting
		// the server may update after (and change if necessary)
		m.buffers[serverKey].Chat.SetNickname(server.Nickname)

		for _, channel := range server.Channels {
			key := makeBufferKey(server.Name, channel)
			buffer := &Buffer{
				Key:    key,
				Server: server.Name,
				Chat:   chat.New(m.theme, m.usernameColors),
				Users:  users.New(m.theme, m.usernameColors),
			}

			buffer.Chat.SetNickname(server.Nickname)
			m.buffers[key] = buffer
		}
	}

	// Make the first server buffer active on startup
	if len(cfg.Servers) > 0 {
		m.activeBuffer = makeBufferKey(cfg.Servers[0].Name, "")
	}

	return &m
}

func (m *Model) Init() tea.Cmd {
	buf := m.getActiveBuffer()
	return buf.Chat.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if sel, ok := msg.(channels.ChannelSelectionMsg); ok && sel.Server != "" {
			key := makeBufferKey(sel.Server, sel.Channel)
			m.activeBuffer = key
			buf := m.getActiveBuffer()
			buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
			buf.Users = buf.Users.SetSize(usersPanelMaxWidth, m.calculateChatHeight())
		}

	case tea.WindowSizeMsg:
		m.width = msgTyped.Width
		m.height = msgTyped.Height

		// TODO: ideally palette width should be smaller than chat width. It might happen if the font size is too big
		m.palette.SetSize(90, len(m.palette.Commands())+5)

		buf := m.getActiveBuffer()
		buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
		buf.Users = buf.Users.SetSize(usersPanelMaxWidth, m.calculateChatHeight())

	case irc.UserListMsg:
		key := makeBufferKey(msgTyped.Server, msgTyped.Channel)
		if buf, ok := m.buffers[key]; ok {
			buf.Users, cmd = buf.Users.Update(users.UserListMsg(msgTyped.Users))
			cmds = append(cmds, cmd)
		}

	case irc.BufferNewMessageMsg:
		key := makeBufferKey(msgTyped.Server, msgTyped.Channel)
		if buf, ok := m.buffers[key]; ok {
			buf.Chat.AddMessage(chat.Message{
				Time:     msgTyped.Time,
				Username: msgTyped.From,
				Text:     msgTyped.Text,
			})
		}

	case irc.ChannelTopicMsg:
		key := makeBufferKey(msgTyped.Server, msgTyped.Channel)
		if buf, ok := m.buffers[key]; ok {
			buf.Chat.SetTopic(msgTyped.Topic)
		}

	case irc.NickUpdateMsg:
		for _, buf := range m.buffers {
			if buf.Server == msgTyped.Server {
				buf.Chat.SetNickname(msgTyped.Nick)
			}
			buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
		}

	case irc.ChannelNameUpdateMsg:
		m.channels, cmd = m.channels.Update(channels.ChannelNameUpdateMsg{
			Server:        msgTyped.Server,
			CanonicalName: msgTyped.CanonicalName,
		})
		cmds = append(cmds, cmd)
	}

	// Write characters in the palette input bar if it's open, instead of chat input
	if m.palette.IsVisible() {
		palModel, cmd := m.palette.Update(msg)
		m.palette = palModel.(*palette.Model)
		cmds = append(cmds, cmd)
	}

	// Update background components, blocking keyboard input when the palette is open
	msg = m.filterOutKeyMsgs(msg)
	buf := m.getActiveBuffer()
	buf.Chat, cmd = buf.Chat.Update(msg)
	cmds = append(cmds, cmd)

	buf.Users, cmd = buf.Users.Update(msg)
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

func (m *Model) View() string {
	if m.palette.IsVisible() && m.overlay != nil {
		return m.overlay.View()
	}

	return (&background{m}).View()
}

// filterOutKeyMsgs filters out keyboard messages from background components when palette is open
func (m *Model) filterOutKeyMsgs(msg tea.Msg) tea.Msg {
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

func (m *Model) getActiveBuffer() *Buffer {
	buf := m.buffers[m.activeBuffer]
	if buf == nil {
		log.Fatalf("Buffer %s not found", m.activeBuffer)
	}
	return buf
}

func (m *Model) calculateChatWidth() int {
	return m.width - channelsPanelMaxWidth - usersPanelMaxWidth - appVerticalBordersSize
}

func (m *Model) calculateChatHeight() int {
	h := m.height - topPadding
	if h < 0 {
		return 0
	}
	return h
}
