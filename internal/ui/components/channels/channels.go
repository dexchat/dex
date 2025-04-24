package channels

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Model struct {
	tree     *tree.Tree
	Selected int

	theme styles.Theme
}

func New(theme styles.Theme) Model {
	servers := map[string][]string{
		"irc.libera.chat": {"#libera", "#go", "#rust"},
		"irc.rizon.net":   {"#kubernetes", "#linux", "#nginx"},
	}

	var serverTree []string
	for server, channels := range servers {
		sub := tree.Root(server)
		for _, channel := range channels {
			sub.Child(channel)
		}

		serverTree = append(serverTree, sub.String())
	}

	t := tree.Root(strings.Join(serverTree, "\n\n"))

	return Model{
		Selected: 0,
		theme:    theme,
		tree:     t,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(width int, height int) string {
	return m.theme.Styles.Sidebar.
		Width(width).
		Height(height).
		Render(m.tree.String())
}
