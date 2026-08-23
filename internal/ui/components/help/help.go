package help

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	helpModalWidth = 48
	boxPadding     = 1
	keyColWidth    = 16
)

type Model struct {
	theme    styles.Theme
	width    int
	height   int
	visible  bool
	viewport viewport.Model
}

type item struct {
	key  string
	desc string
}

func defaultShortcuts() []item {
	kb := keybindings.DefaultKeyMap()
	editInEditor := kb.EditInEditor.Help()

	return []item{
		{key: bindingKeys(kb.TogglePalette), desc: "toggle command palette"},
		{key: bindingKeys(kb.GoToChannel), desc: "channel picker"},
		{key: bindingKeys(kb.LastBuffer), desc: "switch to last channel"},
		{key: bindingHelpGroups(kb.MoveUp, kb.MoveDown), desc: "navigate channels"},
		{key: bindingHelpGroups(kb.ScrollUp, kb.ScrollDown), desc: "scroll chat history"},
		{key: bindingKeys(kb.Autocomplete), desc: "autocomplete nickname"},
		{key: editInEditor.Key, desc: editInEditor.Desc},
		{key: bindingKeys(kb.Quit), desc: "quit dex (press twice)"},
	}
}

func bindingKeys(binding key.Binding) string {
	return strings.Join(binding.Keys(), ", ")
}

func bindingHelpGroups(bindings ...key.Binding) string {
	groups := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		groups = append(groups, binding.Help().Key)
	}
	return strings.Join(groups, " / ")
}

func New(theme styles.Theme) *Model {
	vp := viewport.New(viewport.WithWidth(helpModalWidth-boxPadding*2), viewport.WithHeight(0))
	m := &Model{
		theme:    theme,
		viewport: vp,
		visible:  false,
	}
	m.updateViewportContent()
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEscape, tea.KeyEnter:
			m.Close()
			return m, nil
		case 'q', 'Q':
			m.Close()
			return m, nil
		}

		switch keyMsg.String() {
		case "q", "Q":
			m.Close()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if !m.visible {
		return ""
	}

	layoutWidth := min(m.width, helpModalWidth)
	contentWidth := layoutWidth - boxPadding*2
	if contentWidth < 10 {
		contentWidth = 10
	}

	bgStyle := lipgloss.NewStyle().Background(m.theme.Colors.Base.Background)

	titleStyle := bgStyle.
		Foreground(m.theme.Colors.Base.Accent).
		Bold(true).
		Width(contentWidth).
		Align(lipgloss.Center)

	title := titleStyle.Render("Dex Keybindings")

	footerStyle := bgStyle.
		Foreground(m.theme.Colors.Base.Dimmed).
		Width(contentWidth).
		Align(lipgloss.Center)

	footer := footerStyle.Render("Press Esc, Enter, or q to close")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		m.viewport.View(),
		"",
		footer,
	)

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderBackground(m.theme.Colors.Base.Background).
		BorderForeground(m.theme.Colors.Palette.Border).
		Background(m.theme.Colors.Base.Background).
		Width(layoutWidth).
		Padding(boxPadding)

	return helpBox.Render(content)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	layoutWidth := min(width, helpModalWidth)
	contentWidth := layoutWidth - boxPadding*2
	if contentWidth < 10 {
		contentWidth = 10
	}

	shortcuts := defaultShortcuts()
	contentLines := len(shortcuts)
	maxAvail := max(3, height-8)
	vpHeight := min(contentLines, maxAvail)

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(vpHeight)
	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	contentWidth := m.viewport.Width()
	if contentWidth < 10 {
		contentWidth = helpModalWidth - boxPadding*2
	}

	bgStyle := lipgloss.NewStyle().Background(m.theme.Colors.Base.Background)
	keyStyle := bgStyle.
		Foreground(m.theme.Colors.Base.Foreground).
		Bold(true).
		Width(keyColWidth)

	descWidth := max(1, contentWidth-keyColWidth)
	descStyle := bgStyle.
		Foreground(m.theme.Colors.Base.Dimmed)

	var lines []string
	for _, it := range defaultShortcuts() {
		k := keyStyle.Render(it.key)
		d := descStyle.Width(descWidth).Render(it.desc)
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, k, d))
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func (m *Model) Show() {
	m.visible = true
	m.viewport.GotoTop()
}

func (m *Model) Close() {
	m.visible = false
}

func (m *Model) Toggle() {
	if m.visible {
		m.Close()
	} else {
		m.Show()
	}
}

func (m *Model) IsVisible() bool {
	return m.visible
}
