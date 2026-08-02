package ui

import (
	"log"
	"time"

	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/history"
	"github.com/vaaleyard/dex/internal/irc"
	"github.com/vaaleyard/dex/internal/ui/styles"

	tea "charm.land/bubbletea/v2"
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
	// channels right (1) + users left (1) + chat borders (2)
	appVerticalBordersSize = 4
)

type scrollPane int

const (
	scrollPaneNone scrollPane = iota
	scrollPaneChannels
	scrollPaneChat
	scrollPaneUsers
)

type (
	historyFlushMsg     struct{}
	startConnectionMsg  struct{}
	historyLoadedMsg    []loadedBufferHistory
	loadedBufferHistory struct {
		key     BufferKey
		history *history.Log
	}
)

type Model struct {
	config           *config.Config
	ircClientManager *irc.ClientManager

	width          int
	height         int
	theme          styles.Theme
	usernameColors styles.UsernameColors

	buffers      map[BufferKey]*Buffer
	activeBuffer BufferKey

	channels     channels.Model
	palette      *palette.Model
	flushPending bool
}

func New(cfg *config.Config) *Model {
	m := Model{
		config:  cfg,
		buffers: make(map[BufferKey]*Buffer),
		theme:   styles.AyuDarkTheme(),
	}
	m.usernameColors = styles.NewUsernameColors(m.theme.Colors.Nicknames)

	m.channels = channels.New(m.theme, m.config.Servers)
	m.palette = palette.New(m.theme)

	for _, server := range cfg.Servers {
		serverKey := makeBufferKey(server.Name, "")
		serverBuf := &Buffer{
			Key:     serverKey,
			Server:  server.Name,
			Chat:    chat.New(m.theme, m.usernameColors),
			Users:   users.New(m.theme, m.usernameColors),
			History: history.NewLog(),
		}
		// Set the nickname configured in the config file before connecting;
		// the server may update after (and change if necessary)
		serverBuf.Chat.SetNickname(server.Nickname)
		m.buffers[serverKey] = serverBuf

		for _, channel := range server.Channels {
			key := makeBufferKey(server.Name, channel)
			buffer := &Buffer{
				Key:     key,
				Server:  server.Name,
				Chat:    chat.New(m.theme, m.usernameColors),
				Buffer:  channel,
				Users:   users.New(m.theme, m.usernameColors),
				History: history.NewLog(),
			}

			// Although the Nickname is set per server, it is
			// being set both in server and buffer. The Chat
			// model doesn't access the config to fetch it, and each
			// chat buffer needs to display a nickname in the input bar
			buffer.Chat.SetNickname(server.Nickname)

			buffer.Chat.SetChannelMembers(&buffer.Users)

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
	return tea.Batch(
		buf.Chat.Init(),
		m.scheduleHistoryFlush(),
		m.loadConfiguredHistory(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msgTyped := msg.(type) {
	case tea.KeyPressMsg:
		if msgTyped.String() == "ctrl+c" {
			if keybindings.QuitHandler() {
				m.flushAllHistory()
				return m, tea.Quit
			}
			return m, nil
		}
		msg = m.handleKeybindings(msgTyped)
		if sel, ok := msg.(channels.ChannelSelectionMsg); ok && sel.Server != "" {
			key := makeBufferKey(sel.Server, sel.Channel)
			m.activeBuffer = key
			m.clearBufferActivity(key)
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
		m.channels = m.channels.SetSize(channelsPanelMaxWidth, m.calculateChatHeight())

	case irc.ChannelJoinedMsg:
		_, createCmd := m.getOrCreateBuffer(msgTyped.Server, msgTyped.Channel)
		if createCmd != nil {
			cmds = append(cmds, createCmd)
		}

	case irc.UserListMsg:
		buf, createCmd := m.getOrCreateBuffer(msgTyped.Server, msgTyped.Channel)
		if createCmd != nil {
			cmds = append(cmds, createCmd)
		}
		if buf != nil {
			buf.Users, cmd = buf.Users.Update(users.UserListMsg(msgTyped.Users))
			cmds = append(cmds, cmd)
			// We refresh chat on UserListMsg to dim nick if a user
			// send a message then leaves channel
			if buf.Key == m.activeBuffer {
				buf.Chat.RefreshContent()
			}
		}

	case irc.BufferNewMessageBatchMsg:
		for _, msg := range msgTyped {
			if newBufCmd := m.processIncomingMessage(msg); newBufCmd != nil {
				cmds = append(cmds, newBufCmd)
			}
		}
		if cmd := m.scheduleFlush(); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case irc.BufferNewMessageMsg:
		if newBufCmd := m.processIncomingMessage(msgTyped); newBufCmd != nil {
			cmds = append(cmds, newBufCmd)
		}
		if cmd := m.scheduleFlush(); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case irc.ChannelTopicMsg:
		buf, createCmd := m.getOrCreateBuffer(msgTyped.Server, msgTyped.Channel)
		if createCmd != nil {
			cmds = append(cmds, createCmd)
		}
		if buf != nil {
			buf.Chat.SetTopic(msgTyped.Topic)
			// In case the channel topic is more than one line
			buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
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

	case chat.SendMessageMsg:
		buffer := m.getActiveBuffer()
		if m.ircClientManager != nil && buffer.isValid() {
			now := time.Now()
			buffer.Chat.AddMessage(chat.Message{
				Timestamp: now,
				Username:  buffer.Chat.Nickname(),
				Text:      msgTyped.Text,
			})

			// Don't insert into history here - let the echo-message or playback
			// insert it with the correct server timestamp. We already display
			// it immediately in the UI above.

			// Send it asynchronously to avoid blocking the UI by girc
			go m.ircClientManager.Send(buffer.Server, buffer.Buffer, msgTyped.Text)
		}

	case historyFlushMsg:
		for _, buf := range m.buffers {
			if err := buf.History.Flush(buf.Server, buf.Buffer); err != nil {
				log.Printf("Failed to flush history for %s/%s: %v", buf.Server, buf.Buffer, err)
			}
		}
		cmds = append(cmds, m.scheduleHistoryFlush())
	case startConnectionMsg:
		if m.ircClientManager != nil {
			m.ircClientManager.ConnectAll()
		}

	case historyLoadedMsg:
		for _, loaded := range msgTyped {
			buf := m.buffers[loaded.key]
			if buf == nil {
				continue
			}
			buf.History = loaded.history
			buf.LoadHistory()
			if buf.Key == m.activeBuffer {
				buf.Chat.FlushQueue()
			}
		}
		cmds = append(cmds, func() tea.Msg { return startConnectionMsg{} })

	case flushChatMsg:
		m.flushAllChats()
	}

	// Write characters in the palette input bar if it's open, instead of chat input
	if m.palette.IsVisible() {
		palModel, cmd := m.palette.Update(msg)
		m.palette = palModel
		cmds = append(cmds, cmd)
		if _, ok := msg.(tea.MouseWheelMsg); ok {
			return m, tea.Batch(cmds...)
		}
	}

	if _, ok := msg.(tea.MouseWheelMsg); ok {
		cmds = append(cmds, m.updateHoveredScrollPane(msg))
		return m, tea.Batch(cmds...)
	}

	// Update layout components, blocking keyboard input when the palette is open
	msg = m.filterOutKeyMsgs(msg)
	buf := m.getActiveBuffer()
	buf.Chat, cmd = buf.Chat.Update(msg)
	cmds = append(cmds, cmd)

	buf.Users, cmd = buf.Users.Update(msg)
	cmds = append(cmds, cmd)

	m.channels, cmd = m.channels.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateHoveredScrollPane(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	buf := m.getActiveBuffer()

	switch paneForMouseWheel(msg, m.width) {
	case scrollPaneChannels:
		m.channels, cmd = m.channels.Update(msg)
	case scrollPaneChat:
		buf.Chat, cmd = buf.Chat.Update(msg)
	case scrollPaneUsers:
		buf.Users, cmd = buf.Users.Update(msg)
	}

	return cmd
}

func paneForMouseWheel(msg tea.Msg, width int) scrollPane {
	mouseMsg, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return scrollPaneNone
	}

	if mouseMsg.X < 0 || mouseMsg.X >= width {
		return scrollPaneNone
	}

	if mouseMsg.X < channelsPanelMaxWidth {
		return scrollPaneChannels
	}

	if mouseMsg.X >= width-usersPanelMaxWidth {
		return scrollPaneUsers
	}

	return scrollPaneChat
}

func (m *Model) View() tea.View {
	content := (&layout{m}).View()
	if m.palette.IsVisible() {
		content = overlayCenter(m.palette.View(), content)
	}

	view := tea.NewView(content)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

// filterOutKeyMsgs filters out keyboard messages from layout components when palette is open
func (m *Model) filterOutKeyMsgs(msg tea.Msg) tea.Msg {
	// If palette is not visible, pass all messages through
	if !m.palette.IsVisible() {
		return msg
	}

	// If palette is visible, block keyboard input to layout components
	if _, isKeyMsg := msg.(tea.KeyPressMsg); isKeyMsg {
		return nil
	}

	// Allow non-keyboard messages (e.g., WindowSizeMsg) to reach layout components
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
	if m.height < 0 {
		return 0
	}
	return m.height
}

func (m *Model) SetManager(manager *irc.ClientManager) {
	m.ircClientManager = manager
}

// getOrCreateBuffer returns the buffer for the given server/channel, creates if needed
func (m *Model) getOrCreateBuffer(server, channel string) (*Buffer, tea.Cmd) {
	key := makeBufferKey(server, channel)
	if buf, ok := m.buffers[key]; ok {
		return buf, nil
	}

	if channel == "" {
		return nil, nil
	}

	buf := &Buffer{
		Key:     key,
		Server:  server,
		Buffer:  channel,
		Chat:    chat.New(m.theme, m.usernameColors),
		Users:   users.New(m.theme, m.usernameColors),
		History: history.Load(server, channel),
	}

	// copy nickname from the server buffer
	serverKey := makeBufferKey(server, "")
	if serverBuf, exists := m.buffers[serverKey]; exists {
		buf.Chat.SetNickname(serverBuf.Chat.Nickname())
	}

	buf.Chat.SetChannelMembers(&buf.Users)

	m.buffers[key] = buf

	buf.LoadHistory()

	return buf, func() tea.Msg {
		return channels.NewBufferMsg{
			Server: server,
			Buffer: channel,
		}
	}
}

func (m *Model) loadConfiguredHistory() tea.Cmd {
	type target struct {
		key     BufferKey
		server  string
		channel string
	}

	targets := make([]target, 0, len(m.buffers))
	for _, server := range m.config.Servers {
		targets = append(targets, target{
			key:    makeBufferKey(server.Name, ""),
			server: server.Name,
		})
		for _, channel := range server.Channels {
			targets = append(targets, target{
				key:     makeBufferKey(server.Name, channel),
				server:  server.Name,
				channel: channel,
			})
		}
	}

	return func() tea.Msg {
		loaded := make(historyLoadedMsg, 0, len(targets))
		for _, target := range targets {
			loaded = append(loaded, loadedBufferHistory{
				key:     target.key,
				history: history.Load(target.server, target.channel),
			})
		}
		return loaded
	}
}

func (m *Model) scheduleHistoryFlush() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return historyFlushMsg{}
	})
}

func (m *Model) flushAllHistory() {
	for _, buf := range m.buffers {
		if err := buf.History.Flush(buf.Server, buf.Buffer); err != nil {
			log.Printf("Failed to flush history for %s/%s: %v", buf.Server, buf.Buffer, err)
		}
	}
}
