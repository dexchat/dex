package palette

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	nameMaxLen                   = 18
	keyMaxLen                    = 10
	commandPaletteBoxPaddingSize = 1
	actionLinePaddingSize        = 1
)

type Model struct {
	theme   styles.Theme
	actions []action
	cursor  int
	width   int
	height  int
	visible bool
	input   textinput.Model
	keyMap  keybindings.KeyMap
}

func New(theme styles.Theme) *Model {
	input := textinput.New()
	input.Prompt = "> "
	inputStyle := lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Foreground(theme.Colors.Base.Foreground)
	inputStyles := input.Styles()
	inputStyles.Focused.Text = inputStyle
	inputStyles.Blurred.Text = inputStyle
	input.SetStyles(inputStyles)

	return &Model{
		theme:   theme,
		actions: defaultActions(),
		visible: false,
		cursor:  0,
		input:   input,
		keyMap:  keybindings.DefaultKeyMap(),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch keyMsg.Code {
		case tea.KeyEscape:
			m.close()
			return m, nil
		case tea.KeyEnter:
			return m, m.executeSelected()
		}

		switch {
		case key.Matches(keyMsg, m.keyMap.MoveUp):
			m.moveUp()
			return m, nil
		case key.Matches(keyMsg, m.keyMap.MoveDown):
			m.moveDown()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if !m.visible {
		return ""
	}

	filtered := m.filteredCommands()

	contentWidth := m.width - commandPaletteBoxPaddingSize*2
	bgStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Base.Background).
		Width(contentWidth)

	input := bgStyle.
		MarginBottom(1).
		Render(m.input.View())

	var components []string
	components = append(components, input)

	for i, cmd := range filtered {
		components = append(components, m.renderCommandLine(i, cmd))
	}

	if len(filtered) == 0 && m.input.Value() != "" {
		noMatchStr := bgStyle.Align(lipgloss.Center).Render("  No matching actions")
		components = append(components, noMatchStr)
	}

	// TODO: set a maximum number of actions to show
	for len(components) < len(m.Commands())+1 {
		components = append(components, bgStyle.Render(""))
	}
	components = components[:len(m.Commands())+1]

	content := lipgloss.JoinVertical(lipgloss.Left, components...)

	commandPaletteBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderBackground(m.theme.Colors.Base.Background).
		BorderForeground(m.theme.Colors.Palette.Border).
		Background(m.theme.Colors.Base.Background).
		Width(m.width).
		Padding(commandPaletteBoxPaddingSize)

	return commandPaletteBox.Render(content)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Toggle() {
	m.visible = !m.visible
	if m.visible {
		m.open()
	} else {
		m.close()
	}
}

func (m *Model) IsVisible() bool {
	return m.visible
}

func (m *Model) Commands() []action {
	return m.actions
}

func (m *Model) open() {
	m.cursor = 0
	m.input.Reset()
	m.input.Focus()
}

func (m *Model) close() {
	m.visible = false
	m.cursor = 0
	m.input.Reset()
}

func (m *Model) moveUp() {
	filtered := m.filteredCommands()
	if len(filtered) == 0 {
		return
	}
	m.cursor--
	if m.cursor < 0 {
		m.cursor = len(filtered) - 1
	}
}

func (m *Model) moveDown() {
	filtered := m.filteredCommands()
	if len(filtered) == 0 {
		return
	}
	m.cursor++
	if m.cursor >= len(filtered) {
		m.cursor = 0
	}
}

func (m *Model) executeSelected() tea.Cmd {
	filtered := m.filteredCommands()
	if len(filtered) == 0 || m.cursor >= len(filtered) {
		return nil
	}

	handler := filtered[m.cursor].Handler
	m.close()

	if handler != nil {
		return func() tea.Msg { return handler() }
	}
	return nil
}

func (m *Model) filteredCommands() []action {
	query := m.input.Value()
	if query == "" {
		return m.actions
	}

	var filtered []action
	for _, cmd := range m.actions {
		if cmd.matches(query) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

func (m *Model) renderCommandLine(index int, cmd action) string {
	highlight := index == m.cursor

	// Determine styles based on highlight state
	backgroundColor := m.theme.Colors.Base.Background
	textColor := m.theme.Colors.Base.Foreground
	accentColor := m.theme.Colors.Base.Accent
	if highlight {
		backgroundColor = m.theme.Colors.Palette.Highlight
		textColor = m.theme.Colors.Base.Background
		accentColor = m.theme.Colors.Base.Background
	}

	descWidth := m.width - commandPaletteBoxPaddingSize*2 - nameMaxLen - actionLinePaddingSize*2 - keyMaxLen
	if descWidth < 10 {
		descWidth = 10
	}

	nameStr := lipgloss.NewStyle().
		Background(backgroundColor).
		Foreground(accentColor).
		Width(nameMaxLen).
		Render(cmd.Name)
	descStr := lipgloss.NewStyle().
		Background(backgroundColor).
		Foreground(textColor).
		Width(descWidth).
		Render(cmd.Description)
	keyStr := lipgloss.NewStyle().
		Background(backgroundColor).
		Foreground(m.theme.Colors.Base.Dimmed).
		Width(keyMaxLen).
		Render(cmd.Keybinding.Help().Key)

	// Join columns horizontally
	actionLine := lipgloss.JoinHorizontal(lipgloss.Top, nameStr, descStr, keyStr)

	return lipgloss.NewStyle().
		Background(backgroundColor).
		Padding(0, actionLinePaddingSize).
		Render(actionLine)
}
