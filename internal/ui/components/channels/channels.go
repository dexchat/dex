package channels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type node struct {
	name         string
	isServer     bool
	parent       string
	mentioned    bool
	mentionCount int
	hasUnread    bool
}

type Model struct {
	nodes    []node
	cursor   int
	selected string

	theme styles.Theme
}

func New(theme styles.Theme) Model {
	// hardcoded only for POC, the logic will be implemented later
	servers := map[string][]string{
		"irc.libera.chat": {"#libera", "#go", "#rust", "#homelab", "#ai", "#go-nuts", "#linux", "#kde", "#archlinux",
			"#python", "#debian", "##thelouge", "#ubuntu"},
		"irc.rizon.net": {"#kubernetes", "#linux", "#nginx", "#go"},
	}

	var items []node

	// server:channel differentiates channels with the same name in different servers
	mentionCounts := map[string]int{
		"irc.libera.chat:#go": 2,
	}
	unreadChannels := map[string]bool{
		"irc.libera.chat:#libera": true,
		"irc.libera.chat:#debian": true,
	}

	for server, channels := range servers {
		items = append(items, node{
			name:     server,
			isServer: true,
		})

		for _, channel := range channels {
			key := server + ":" + channel
			isMentioned := mentionCounts[key] > 0
			items = append(items, node{
				name:         channel,
				isServer:     false,
				parent:       server,
				mentioned:    isMentioned,
				mentionCount: mentionCounts[key],
				hasUnread:    unreadChannels[key] && !isMentioned,
			})
		}
	}

	return Model{
		cursor:   0,
		selected: "",
		nodes:    items,
		theme:    theme,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) processInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "ctrl+n":
		if m.cursor < len(m.nodes)-1 {
			m.cursor++
		}
	}

	if m.cursor < len(m.nodes) {
		currentItem := m.nodes[m.cursor]
		if !currentItem.isServer {
			m.selected = currentItem.parent + ":" + currentItem.name
			return m, func() tea.Msg {
				return ChannelSelectedMsg{
					Channel: currentItem.name,
					Server:  currentItem.parent,
				}
			}
		}
	}

	return m, nil
}

type ChannelSelectedMsg struct {
	Channel string
	Server  string
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.processInput(msg)
	}
	return m, nil
}

func (m Model) View(width int, height int) string {
	var lines []string
	lastServerName := ""

	for i, node := range m.nodes {
		if node.isServer && lastServerName != "" {
			lines = append(lines, "")
		}

		var line string
		var textPart string

		if node.isServer {
			textPart = node.name
			lastServerName = node.name
		} else {
			channelText := node.name
			if node.mentioned && node.mentionCount > 0 {
				channelText = fmt.Sprintf("%s (%d)", node.name, node.mentionCount)
			}
			textPart = channelText
		}

		var itemStyle = m.theme.Styles.App.PaddingLeft(1).PaddingRight(1)

		if node.isServer {
			itemStyle = m.theme.Styles.ServerItem.PaddingLeft(1).PaddingRight(1)
		} else if node.mentioned {
			itemStyle = m.theme.Styles.MentionedItem.PaddingLeft(1).PaddingRight(1)
		} else if node.hasUnread {
			itemStyle = m.theme.Styles.UnreadItem.PaddingLeft(1).PaddingRight(1)
		}

		if i == m.cursor {
			itemStyle = itemStyle.Background(m.theme.Colors.BorderColor)
		}

		styledText := itemStyle.Render(textPart)
		if node.isServer {
			line = styledText
		} else {
			line = fmt.Sprintf("  %s", styledText)
		}

		lines = append(lines, line)
	}

	channelsBody := strings.Join(lines, "\n")
	helpView := renderHelp(width, m.theme)

	// help should be placed at one line above the bottom margin
	spacingLines := height - len(lines) - 1
	if spacingLines < 0 {
		spacingLines = 0
	}
	body := channelsBody + "\n" + strings.Repeat("\n", spacingLines) + helpView

	return m.theme.Styles.Sidebar.
		BorderRightForeground(m.theme.Colors.LighterBackground).
		Width(width).
		Height(height).
		Render(body)
}
