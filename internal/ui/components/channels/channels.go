package channels

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type node struct {
	name     string
	isServer bool
	parent   string
}

type Model struct {
	rawTree  string
	nodes    []node
	cursor   int
	selected string

	theme styles.Theme
}

func New(theme styles.Theme) Model {
	servers := map[string][]string{
		"irc.libera.chat": {"#libera", "#go", "#rust"},
		"irc.rizon.net":   {"#kubernetes", "#linux", "#nginx"},
	}

	var serverTree []string
	var items []node

	// Build the tree and item list once
	for server, channels := range servers {
		sub := tree.Root(server)
		items = append(items, node{name: server, isServer: true})

		for _, channel := range channels {
			sub.Child(channel)
			items = append(items, node{name: channel, isServer: false, parent: server})
		}

		serverTree = append(serverTree, sub.String())
	}

	t := tree.Root(strings.Join(serverTree, "\n\n"))

	return Model{
		cursor:   0,
		selected: "",
		nodes:    items,
		rawTree:  t.String(),
		theme:    theme,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) ProcessInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.nodes)-1 {
			m.cursor++
		}
	case "enter":
		// Only allow selecting channels, not servers
		currentItem := m.nodes[m.cursor]
		if !currentItem.isServer {
			m.selected = currentItem.name
			return m, func() tea.Msg {
				return ChannelSelectedMsg{
					Channel: m.selected,
					Server:  currentItem.parent,
				}
			}
		}
	}
	return m, nil
}

// Custom message for channel selection
type ChannelSelectedMsg struct {
	Channel string
	Server  string
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.ProcessInput(msg)
	}
	return m, nil
}

func (m Model) View(width int, height int) string {
	// Apply styling to the raw tree string based on cursor position and selection
	lines := strings.Split(m.rawTree, "\n")
	styledLines := make([]string, 0, len(lines))

	currentItemIndex := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			styledLines = append(styledLines, line)
			continue
		}

		// Check if this line corresponds to an item
		isItemLine := false
		for _, item := range []string{"├──", "└──", "│"} {
			if strings.Contains(line, item) {
				isItemLine = true
				break
			}
		}

		if isItemLine || !strings.HasPrefix(line, " ") {
			// This is either a server or channel line
			var itemStyle = m.theme.Styles.App

			if currentItemIndex == m.cursor {
				itemStyle = m.theme.Styles.SelectedItem
			}

			// Apply underline for selected channel
			if currentItemIndex < len(m.nodes) &&
				m.nodes[currentItemIndex].name == m.selected {
				itemStyle = itemStyle.Underline(true)
			}

			// Extract the item name from the line
			itemName := m.nodes[currentItemIndex].name

			// Replace the item name with styled version while preserving tree structure
			styledLine := strings.Replace(
				line,
				itemName,
				itemStyle.Render(itemName),
				1,
			)

			styledLines = append(styledLines, styledLine)
			currentItemIndex++
		} else {
			// This is a tree structure line (not an item)
			styledLines = append(styledLines, line)
		}
	}

	return m.theme.Styles.Sidebar.
		Width(width).
		Height(height).
		Render(strings.Join(styledLines, "\n"))
}

// SelectedChannel returns the currently selected channel
func (m Model) SelectedChannel() string {
	return m.selected
}

// SelectedServer returns the server of the currently selected channel
func (m Model) SelectedServer() string {
	for _, item := range m.nodes {
		if item.name == m.selected {
			return item.parent
		}
	}
	return ""
}
