package channels

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	// left (1) + right (1) borders
	channelsVerticalBordersSize = 2
	bottomPaddingHeight         = 1
)

type ChannelSelectionMsg struct {
	Channel string
	Server  string
}

type ChannelNameUpdateMsg struct {
	Server        string
	CanonicalName string
}

type ActivityUpdateMsg struct {
	Server       string
	Buffer       string
	UnreadCount  int
	MentionCount int
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
	mentionCount int
	unreadCount  int
}

type Model struct {
	nodes    []node
	cursor   int
	selected string
	viewport viewport.Model

	theme styles.Theme

	servers []*config.Server
}

func New(theme styles.Theme, servers []*config.Server) Model {
	var items []node

	for _, server := range servers {
		items = append(items, node{
			name:     server.Name,
			isServer: true,
		})

		for _, channel := range server.Channels {
			items = append(items, node{
				name:     channel,
				isServer: false,
				parent:   server.Name,
			})
		}
	}

	vp := viewport.New(
		viewport.WithWidth(0),
		viewport.WithHeight(0),
	)
	vp.Style = lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		PaddingBottom(bottomPaddingHeight)

	return Model{
		cursor:   0,
		selected: "",
		nodes:    items,
		viewport: vp,
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
	case ActivityUpdateMsg:
		for i, n := range m.nodes {
			if !n.isServer && strings.EqualFold(n.parent, msg.Server) && strings.EqualFold(n.name, msg.Buffer) {
				m.nodes[i].unreadCount = msg.UnreadCount
				m.nodes[i].mentionCount = msg.MentionCount
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

	if _, ok := msg.(tea.KeyMsg); ok {
		return m.updateContent(), nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m.updateContent(), cmd
}

func (m Model) SetSize(width, height int) Model {
	contentWidth := width - channelsVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	helpView := renderFooter(contentWidth, m.theme)
	helpHeight := lipgloss.Height(helpView)
	listHeight := height - m.theme.Styles.Sidebar.GetVerticalFrameSize() - helpHeight
	if listHeight < 0 {
		listHeight = 0
	}

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(listHeight)
	return m.updateContent()
}

func (m Model) updateContent() Model {
	var lines []string
	lastServerName := ""

	for i, node := range m.nodes {
		if node.isServer && lastServerName != "" {
			lines = append(lines, "")
		}

		if node.isServer {
			lastServerName = node.name
		}

		var itemStyle = m.theme.Styles.App

		if node.isServer {
			itemStyle = m.theme.Styles.ServerItem
		} else if node.mentionCount > 0 {
			itemStyle = m.theme.Styles.MentionedItem
		} else if node.unreadCount > 0 {
			itemStyle = m.theme.Styles.UnreadItem
		}

		if i == m.cursor {
			itemStyle = itemStyle.Background(m.theme.Colors.Sidebar.Selection)
		}

		itemStyle = itemStyle.Width(m.viewport.Width())

		if node.isServer {
			lines = append(lines, itemStyle.Render(node.name))
		} else {
			lines = append(lines, itemStyle.Render(m.renderChannelRow(node, m.viewport.Width())))
		}
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
	return m
}

func (m Model) renderChannelRow(node node, width int) string {
	label := "  " + node.name
	badge := m.renderActivityBadge(node)
	if badge == "" {
		return label
	}

	available := width - lipgloss.Width(badge)
	if available < 0 {
		available = 0
	}

	label = truncateWidth(label, available)
	spacing := available - lipgloss.Width(label)
	if spacing < 0 {
		spacing = 0
	}

	return label + strings.Repeat(" ", spacing) + badge
}

func (m Model) renderActivityBadge(node node) string {
	switch {
	case node.mentionCount > 0:
		return m.activityBadge("@"+formatActivityCount(node.mentionCount), m.theme.Styles.MentionedItem)
	case node.unreadCount > 0:
		return m.activityBadge(formatActivityCount(node.unreadCount), m.theme.Styles.UnreadItem)
	default:
		return ""
	}
}

func (m Model) activityBadge(text string, style lipgloss.Style) string {
	return style.
		Background(m.theme.Colors.Base.Surface).
		PaddingLeft(1).
		PaddingRight(1).
		Render(text)
}

func formatActivityCount(count int) string {
	if count > 99 {
		return "99+"
	}
	return fmt.Sprintf("%d", count)
}

func truncateWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}

	var out strings.Builder
	for _, r := range s {
		next := out.String() + string(r)
		if lipgloss.Width(next) > width {
			break
		}
		out.WriteRune(r)
	}
	return out.String()
}

func (m Model) View(width, height int) string {
	contentWidth := width - channelsVerticalBordersSize
	if contentWidth < 0 {
		contentWidth = 0
	}

	helpView := renderFooter(contentWidth, m.theme)
	helpHeight := lipgloss.Height(helpView)
	listHeight := height - m.theme.Styles.Sidebar.GetVerticalFrameSize() - helpHeight
	if listHeight < 0 {
		listHeight = 0
	}

	if m.viewport.Width() != contentWidth || m.viewport.Height() != listHeight {
		m = m.SetSize(width, height)
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
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
