package channels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	// left (1) + right (1) borders
	channelsVerticalBordersSize = 2
)

type ChannelSelectionMsg struct {
	Channel string
	Server  string
}

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

	servers map[string]*config.Server
}

func New(theme styles.Theme, servers map[string]*config.Server) Model {
	var items []node

	// example: "libera:#go": 2
	mentionCounts := map[string]int{}

	// example: "libera:#go": true
	unreadChannels := map[string]bool{}

	for serverName, server := range servers {
		items = append(items, node{
			name:     serverName,
			isServer: true,
		})

		for _, channel := range server.Channels {
			key := serverName + ":" + channel
			isMentioned := mentionCounts[key] > 0
			items = append(items, node{
				name:         channel,
				isServer:     false,
				parent:       serverName,
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
		servers:  servers,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(width, height int) string {
	contentWidth := width - channelsVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

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
	helpView := renderFooter(contentWidth, m.theme)
	helpHeight := lipgloss.Height(helpView)

	listHeight := height - helpHeight
	if listHeight < 0 {
		listHeight = 0
	}

	channelsList := lipgloss.NewStyle().
		Background(m.theme.Colors.Background).
		Width(contentWidth).
		Height(listHeight).
		Render(channelsBody)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		channelsList,
		helpView,
	)

	return m.theme.Styles.Sidebar.
		BorderRightForeground(m.theme.Colors.LighterBackground).
		Height(height).
		Render(body)
}

// MoveUp moves the cursor up to the previous channel
func (m Model) MoveUp() Model {
	if m.cursor > 0 {
		m.cursor--
	}
	if m.cursor < len(m.nodes) {
		currentItem := m.nodes[m.cursor]
		if !currentItem.isServer {
			m.selected = currentItem.parent + ":" + currentItem.name
		}
	}
	return m
}

// MoveDown moves the cursor down to the next channel
func (m Model) MoveDown() Model {
	if m.cursor < len(m.nodes)-1 {
		m.cursor++
	}
	if m.cursor < len(m.nodes) {
		currentItem := m.nodes[m.cursor]
		if !currentItem.isServer {
			m.selected = currentItem.parent + ":" + currentItem.name
		}
	}
	return m
}

func (m Model) Selected() ChannelSelectionMsg {
	currentItem := m.nodes[m.cursor]
	if currentItem.isServer {
		return ChannelSelectionMsg{
			Server: currentItem.name,
		}
	}
	return ChannelSelectionMsg{
		Channel: currentItem.name,
		Server:  currentItem.parent,
	}
}
