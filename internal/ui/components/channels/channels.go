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

type ChannelNameUpdateMsg struct {
	Server        string
	CanonicalName string
}

// NewBufferMsg is a custom message type used to notify this component to add a new buffer in the tree
type NewBufferMsg struct {
	Server string
	Buffer string
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

	servers []*config.Server
}

func New(theme styles.Theme, servers []*config.Server) Model {
	var items []node

	// example: "libera:#go": 2
	mentionCounts := map[string]int{}

	// example: "libera:#go": true
	unreadChannels := map[string]bool{}

	for _, server := range servers {
		items = append(items, node{
			name:     server.Name,
			isServer: true,
		})

		for _, channel := range server.Channels {
			key := server.Name + ":" + channel
			isMentioned := mentionCounts[key] > 0
			items = append(items, node{
				name:         channel,
				isServer:     false,
				parent:       server.Name,
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
	switch msg := msg.(type) {
	case ChannelNameUpdateMsg:
		for i, n := range m.nodes {
			if !n.isServer && n.parent == msg.Server && strings.EqualFold(n.name, msg.CanonicalName) {
				m.nodes[i].name = msg.CanonicalName
				break
			}
		}
	case NewBufferMsg:
		// find the position to insert: after the server and its existing channels
		insertIdx := len(m.nodes) // default to end
		for i, existingNode := range m.nodes {
			if existingNode.isServer && existingNode.name == msg.Server {
				// found the server, now find where its channels end
				insertIdx = i + 1
				for j := i + 1; j < len(m.nodes); j++ {
					if m.nodes[j].isServer {
						break // hit the next server
					}
					insertIdx = j + 1
				}
				break
			}
		}
		newNode := node{
			name:     msg.Buffer,
			isServer: false,
			parent:   msg.Server,
		}
		// insert at the correct position
		m.nodes = append(m.nodes[:insertIdx], append([]node{newNode}, m.nodes[insertIdx:]...)...)
	}
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
				channelText = fmt.Sprintf("%s (%d)", channelText, node.mentionCount)
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
			itemStyle = itemStyle.Background(m.theme.Colors.Sidebar.Selection)
		}

		if node.isServer {
			line = itemStyle.Render(textPart)
		} else {
			line = itemStyle.PaddingLeft(2).Render(textPart)
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
		Background(m.theme.Colors.Base.Background).
		Width(contentWidth).
		Height(listHeight).
		Render(channelsBody)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		channelsList,
		helpView,
	)

	return m.theme.Styles.Sidebar.
		BorderRightForeground(m.theme.Colors.Base.Surface).
		Height(height).
		Render(body)
}

// MoveUp moves the cursor up to the previous channel
func (m Model) MoveUp() Model {
	if len(m.nodes) == 0 {
		return m
	}

	if m.cursor > 0 {
		m.cursor--
	} else {
		m.cursor = len(m.nodes) - 1
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
	if len(m.nodes) == 0 {
		return m
	}

	if m.cursor < len(m.nodes)-1 {
		m.cursor++
	} else {
		m.cursor = 0
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
