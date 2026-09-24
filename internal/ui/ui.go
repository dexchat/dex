package ui

import (
	"strings"
	"time"

	"github.com/dexchat/dex/internal/commands"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/history"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/ui/styles"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/ui/components/channels"
	"github.com/dexchat/dex/internal/ui/components/chat"
	"github.com/dexchat/dex/internal/ui/components/help"
	"github.com/dexchat/dex/internal/ui/components/keybindings"
	"github.com/dexchat/dex/internal/ui/components/palette"
	"github.com/dexchat/dex/internal/ui/components/users"
)

const (
	// Arbitrary values that I think looks good in the screen
	// I believe that having a slightly bigger chat panel compared to
	// the user panel is better for the eyes
	channelsPanelMaxWidth = 25
	usersPanelMaxWidth    = 20
	commandPaletteWidth   = 90
	paletteScreenMargin   = 4
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
	startConnectionMsg struct{}
)

type Model struct {
	config           *config.Config
	ircClientManager *irc.ClientManager

	width          int
	height         int
	theme          styles.Theme
	usernameColors styles.UsernameColors

	buffers             map[BufferKey]*Buffer
	activeBuffer        BufferKey
	lastBuffer          BufferKey
	readState           *history.ReadState
	readStateDirty      bool
	directMessages      *history.DirectMessages
	directMessagesDirty bool

	channels     channels.Model
	palette      *palette.Model
	help         *help.Model
	flushPending bool

	// ircPending holds events pulled from each server but not yet applied.
	ircPending map[string][]irc.Event

	terminalFocused           bool
	lastSoundAt               time.Time
	notificationNoticeVersion uint64
	editKeyPending            bool
	now                       func() time.Time

	// persistence coordinates asynchronous history saves and shutdown state.
	persistence persistenceState
}

func New(cfg *config.Config) *Model {
	directMessages, err := history.LoadDirectMessages()
	if err != nil {
		directMessages = &history.DirectMessages{}
	}
	theme := styles.RosePineTheme()
	switch cfg.UI.Theme {
	case config.ThemeAyuDark:
		theme = styles.AyuDarkTheme()
	case config.ThemeDracula:
		theme = styles.DraculaTheme()
	case config.ThemeGruvboxDark:
		theme = styles.GruvboxDarkTheme()
	case config.ThemeSolarizedLight:
		theme = styles.SolarizedLightTheme()
	}
	m := Model{
		config:          cfg,
		buffers:         make(map[BufferKey]*Buffer),
		ircPending:      make(map[string][]irc.Event),
		readState:       &history.ReadState{},
		directMessages:  directMessages,
		theme:           theme,
		terminalFocused: true,
		now:             time.Now,
		persistence: persistenceState{
			detachedHistory: make(map[BufferKey]detachedHistory),
		},
	}
	m.usernameColors = styles.NewUsernameColors(m.theme.Colors.Nicknames)

	m.channels = channels.New(m.theme, m.config.Servers)
	m.palette = palette.New(m.theme)
	m.help = help.New(m.theme)

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
		cmd               tea.Cmd
		cmds              []tea.Cmd
		forwardToChildren bool
	)

	switch msgTyped := msg.(type) {
	case tea.KeyPressMsg:
		if msgTyped.String() == "ctrl+c" {
			if keybindings.QuitHandler() {
				return m, m.requestShutdown()
			}
			return m, nil
		}
		msg = m.handleKeybindings(msgTyped)
		if sel, ok := msg.(channels.ChannelSelectionMsg); ok {
			m.selectBuffer(sel.Server, sel.Channel, &cmds)
		}
		if _, ok := msg.(palette.LastBufferMsg); ok {
			m.selectLastBuffer(&cmds)
		}
		if _, ok := msg.(palette.EditInEditorMsg); ok {
			cmds = append(cmds, m.editActiveInput())
		}
		switch msg.(type) {
		case nil, channels.ChannelSelectionMsg, palette.LastBufferMsg, palette.EditInEditorMsg:
		default:
			forwardToChildren = true
		}

	case palette.OpenChannelPickerMsg:
		m.showChannelPicker()

	case palette.OpenHelpMsg:
		m.showHelp()

	case palette.EditInEditorMsg:
		cmds = append(cmds, m.editActiveInput())

	case editorFinishedMsg:
		m.finishEditingInput(msgTyped)

	case palette.LastBufferMsg:
		m.selectLastBuffer(&cmds)

	case palette.ChannelSelectionMsg:
		m.selectBuffer(msgTyped.Server, msgTyped.Channel, &cmds)

	case tea.FocusMsg:
		m.terminalFocused = true
		m.clearBufferActivity(m.activeBuffer)

	case tea.BlurMsg:
		m.terminalFocused = false

	case tea.WindowSizeMsg:
		m.width = msgTyped.Width
		m.height = msgTyped.Height

		paletteWidth := min(commandPaletteWidth, max(1, msgTyped.Width-paletteScreenMargin))
		m.palette.SetSize(paletteWidth, m.calculateChatHeight())
		m.help.SetSize(paletteWidth, m.calculateChatHeight())

		buf := m.getActiveBuffer()
		buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
		buf.Users = buf.Users.SetSize(usersPanelMaxWidth, m.calculateChatHeight())
		m.channels = m.channels.SetSize(channelsPanelMaxWidth, m.calculateChatHeight())

	case ircEventsMsg:
		m.ircPending[msgTyped.server] = append(m.ircPending[msgTyped.server], msgTyped.events...)
		cmds = append(cmds, m.applyPendingIRCEvents(msgTyped.server))

	case ircContinueMsg:
		cmds = append(cmds, m.applyPendingIRCEvents(msgTyped.server))

	case ircCommandFailedMsg:
		m.showIRCCommandError(msgTyped)

	case historyFlushMsg:
		if cmd := m.startPersistence(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, m.scheduleHistoryFlush())
	case historyFlushFinishedMsg:
		if cmd := m.applyPersistenceResult(msgTyped); cmd != nil {
			return m, cmd
		}

	case startConnectionMsg:
		if m.ircClientManager != nil {
			m.ircClientManager.ConnectAll()
			for _, server := range m.config.Servers {
				cmds = append(cmds, m.waitIRCEvents(server.Name))
			}
		}

	case historyLoadedMsg:
		m.applyLoadedHistory(msgTyped)

	case initialHistoryLoadedMsg:
		m.applyLoadedHistory(msgTyped.histories)
		m.readState = msgTyped.readState
		if m.readState == nil {
			m.readState = &history.ReadState{}
		}
		cmds = append(cmds, func() tea.Msg { return startConnectionMsg{} })

	case flushChatMsg:
		m.flushAllChats()

	case notificationNoticeExpiredMsg:
		if msgTyped.version == m.notificationNoticeVersion {
			m.channels = m.channels.ClearNotificationNotice()
		}

	default:
		// Unknown messages may belong to a child Bubbles model. Keep forwarding
		// them so private cursor, viewport, and component messages still work.
		forwardToChildren = true
	}

	if !forwardToChildren {
		return m, tea.Batch(cmds...)
	}

	paletteWasVisible := m.palette.IsVisible()
	helpWasVisible := m.help.IsVisible()

	// Write characters in the palette input bar if it's open, instead of chat input
	if paletteWasVisible {
		palModel, cmd := m.palette.Update(msg)
		m.palette = palModel
		cmds = append(cmds, cmd)
		if _, ok := msg.(tea.MouseWheelMsg); ok {
			return m, tea.Batch(cmds...)
		}
	} else if helpWasVisible {
		helpModel, cmd := m.help.Update(msg)
		m.help = helpModel
		cmds = append(cmds, cmd)
		if _, ok := msg.(tea.MouseWheelMsg); ok {
			return m, tea.Batch(cmds...)
		}
	}

	if _, ok := msg.(tea.MouseWheelMsg); ok {
		cmds = append(cmds, m.updateHoveredScrollPane(msg))
		return m, tea.Batch(cmds...)
	}

	// If palette or help was open, consume keyboard input so it doesn't leak into chat/channels
	if paletteWasVisible || helpWasVisible {
		if _, isKeyMsg := msg.(tea.KeyPressMsg); isKeyMsg {
			return m, tea.Batch(cmds...)
		}
	}

	// Update layout components, blocking keyboard input when the palette or help is open
	msg = m.filterOutKeyMsgs(msg)
	buf := m.getActiveBuffer()
	var action *chat.SendAction
	buf.Chat, action, cmd = buf.Chat.Update(msg)
	cmds = append(cmds, cmd)
	if action != nil {
		cmds = append(cmds, m.handleChatSend(buf, action.Text))
		return m, tea.Batch(cmds...)
	}

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
		buf.Chat, _, cmd = buf.Chat.Update(msg)
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
	} else if m.help.IsVisible() {
		content = overlayCenter(m.help.View(), content)
	}

	view := tea.NewView(content)
	view.AltScreen = true
	view.ReportFocus = true
	view.MouseMode = tea.MouseModeCellMotion
	view.BackgroundColor = m.theme.Colors.Base.Background
	return view
}

// filterOutKeyMsgs filters out keyboard messages from layout components when palette or help is open
func (m *Model) filterOutKeyMsgs(msg tea.Msg) tea.Msg {
	// If palette and help are not visible, pass all messages through
	if !m.palette.IsVisible() && !m.help.IsVisible() {
		return msg
	}

	// If palette or help is visible, block keyboard input to layout components
	if _, isKeyMsg := msg.(tea.KeyPressMsg); isKeyMsg {
		return nil
	}

	// Allow non-keyboard messages (e.g., WindowSizeMsg) to reach layout components
	return msg
}

func (m *Model) getActiveBuffer() *Buffer {
	buf := m.buffers[m.activeBuffer]
	if buf == nil {
		panic("active buffer not found: " + string(m.activeBuffer))
	}
	return buf
}

func (m *Model) selectBuffer(server, channel string, cmds *[]tea.Cmd) {
	if server == "" {
		return
	}
	key := makeBufferKey(server, channel)
	buf := m.buffers[key]
	if buf == nil {
		return
	}

	if key != m.activeBuffer {
		m.lastBuffer = m.activeBuffer
		m.activeBuffer = key
	}
	m.clearBufferActivity(key)
	m.channels = m.channels.ClearNotificationNoticeFor(server, channel)
	m.channels, _ = m.channels.Select(server, channel)
	buf.Users = buf.Users.SetSize(usersPanelMaxWidth, m.calculateChatHeight())
	var cmd tea.Cmd
	// Sync membership before SetSize, which reads it while re-rendering.
	buf.Users, cmd = buf.Users.Update(users.UserListMsg{Users: buf.members, Prefixes: buf.memberPrefixes})
	*cmds = append(*cmds, cmd)
	buf.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
	if historyCmd := m.requestHistoryLoad(buf); historyCmd != nil {
		*cmds = append(*cmds, historyCmd)
	}
}

func (m *Model) selectLastBuffer(cmds *[]tea.Cmd) {
	buf := m.buffers[m.lastBuffer]
	if buf == nil {
		return
	}
	m.selectBuffer(buf.Server, buf.Buffer, cmds)
}

func (m *Model) showChannelPicker() {
	channelItems := make([]palette.Channel, 0, len(m.buffers))
	for _, buf := range m.buffers {
		if buf.Buffer != "" {
			channelItems = append(channelItems, palette.Channel{Server: buf.Server, Name: buf.Buffer})
		}
	}
	m.palette.ShowChannels(channelItems)
}

func (m *Model) showHelp() {
	m.palette.Close()
	m.help.SetSize(m.width, m.calculateChatHeight())
	m.help.Show()
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

	log := history.NewLog()
	if detached, ok := m.persistence.detachedHistory[key]; ok {
		log = detached.log
		delete(m.persistence.detachedHistory, key)
	}
	buf := &Buffer{
		Key:     key,
		Server:  server,
		Buffer:  channel,
		Chat:    chat.New(m.theme, m.usernameColors),
		Users:   users.New(m.theme, m.usernameColors),
		History: log,
	}

	// copy nickname from the server buffer
	serverKey := makeBufferKey(server, "")
	if serverBuf, exists := m.buffers[serverKey]; exists {
		buf.Chat.SetNickname(serverBuf.Chat.Nickname())
	}

	// DM buffers have no NAMES-derived member list, so leave channelMembers
	// unset there; otherwise every DM peer would be treated as "not in the
	// channel" and rendered with the inactive/gray nick color.
	if irc.IsChannel(channel) {
		buf.Chat.SetChannelMembers(&buf.Users)
	}

	m.buffers[key] = buf

	newBufferCmd := func() tea.Msg {
		return channels.NewBufferMsg{
			Server: server,
			Buffer: channel,
		}
	}
	return buf, newBufferCmd
}

func (m *Model) restoreDirectMessages(server string) []tea.Cmd {
	var cmds []tea.Cmd
	for _, directMessage := range m.directMessages.Users {
		if !strings.EqualFold(directMessage.Server, server) {
			continue
		}
		_, cmd := m.getOrCreateBuffer(directMessage.Server, directMessage.User)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

func (m *Model) removeChannelBuffer(server, channel string) tea.Cmd {
	key := makeBufferKey(server, channel)
	buf, exists := m.buffers[key]
	if !exists {
		return nil
	}
	return m.removeBuffer(buf, false)
}

func (m *Model) requestHistoryLoad(buf *Buffer) tea.Cmd {
	if buf.historyState == historyLoaded || buf.historyState == historyLoading {
		return nil
	}
	buf.historyState = historyLoading
	return m.loadBufferHistory(buf.Key, buf.Server, buf.Buffer)
}

func (m *Model) loadBufferHistory(key BufferKey, server, channel string) tea.Cmd {
	return func() tea.Msg {
		log, quarantined, err := history.LoadOrQuarantine(server, channel)
		return historyLoadedMsg{{
			key:         key,
			history:     log,
			quarantined: quarantined,
			err:         err,
		}}
	}
}

func (m *Model) applyLoadedHistory(loadedHistory []loadedBufferHistory) {
	for _, loaded := range loadedHistory {
		buf := m.buffers[loaded.key]
		if buf == nil {
			continue
		}
		if loaded.err != nil {
			m.markHistoryUnreadable(buf, loaded.err)
			continue
		}
		loaded.history.Merge(buf.History.Entries())
		buf.History = loaded.history
		buf.historyState = historyLoaded
		buf.LoadHistory()
		if loaded.quarantined != "" {
			m.addCommandError(buf, quarantineNotice(loaded.quarantined))
		}
		if buf.Key == m.activeBuffer {
			buf.Chat.FlushQueue()
		}
	}
}

func (m *Model) loadConfiguredHistory() tea.Cmd {
	return func() tea.Msg {
		readState, err := history.LoadReadState()
		return initialHistoryLoadedMsg{
			readState: readState,
			readErr:   err,
		}
	}
}

func (m *Model) scheduleHistoryFlush() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return historyFlushMsg{}
	})
}

func (m *Model) handleChatSend(buffer *Buffer, text string) tea.Cmd {
	if command, ok := commands.Parse(text); ok {
		return m.handleCommand(buffer, command)
	}
	if m.ircClientManager == nil || !buffer.isValid() {
		return nil
	}
	buffer.Chat.AddMessage(chat.Message{
		Timestamp: time.Now(),
		Username:  buffer.Chat.Nickname(),
		Text:      text,
	})
	// Persist the server echo with its authoritative timestamp and message ID.
	return m.sendMessageCmd(buffer, text)
}
