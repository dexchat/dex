package palette

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

const (
	nameMaxLen = 18
	keyMaxLen  = 10
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
	input.TextStyle = lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Foreground(theme.Colors.Text)

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

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyEsc:
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

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.cursor = 0 // Reset cursor on input query change
	return m, cmd
}

func (m *Model) View() string {
	if !m.visible {
		return ""
	}

	filtered := m.filteredCommands()

	// TODO: move numbers to consts
	contentWidth := m.width - 4      // borders (2) + padding (1) + actionLine horizontal padding (1)
	m.input.Width = contentWidth - 3 // borders (2) + padding (1)
	bgStyle := lipgloss.NewStyle().
		Background(m.theme.Colors.Background).
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
		BorderBackground(m.theme.Colors.Background).
		BorderForeground(m.theme.Colors.Accent).
		Background(m.theme.Colors.Background).
		Width(m.width).
		Padding(1)

	return commandPaletteBox.Render(content)
}

// SetSize sets the dimensions of the palette.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Toggle toggles the visibility of the palette.
func (m *Model) Toggle() {
	m.visible = !m.visible
	if m.visible {
		m.open()
	} else {
		m.close()
	}
}

// IsVisible returns whether the palette is visible.
func (m *Model) IsVisible() bool {
	return m.visible
}

// Commands returns the list of available actions.
func (m *Model) Commands() []action {
	return m.actions
}

// Private helpers

// open prepares the palette for display.
func (m *Model) open() {
	m.cursor = 0
	m.input.Reset()
	m.input.Focus()
}

// close hides the palette and resets state.
func (m *Model) close() {
	m.visible = false
	m.cursor = 0
	m.input.Reset()
}

// moveUp moves the cursor up, wrapping to the bottom if at the top.
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

// moveDown moves the cursor down, wrapping to the top if at the bottom.
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

// executeSelected executes the selected command and closes the palette.
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

// filteredCommands returns the list of actions matching the current input query.
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

// renderCommandLine renders a single line in the list.
func (m *Model) renderCommandLine(index int, cmd action) string {
	highlight := index == m.cursor

	// Determine styles based on highlight state
	backgroundColor := m.theme.Colors.Background
	textColor := m.theme.Colors.Text
	accentColor := m.theme.Colors.Accent
	if highlight {
		backgroundColor = m.theme.Colors.Accent
		textColor = m.theme.Colors.Background
		accentColor = m.theme.Colors.Background
	}

	// TODO: move values to consts
	// Calculate available width (accounting for border and padding)
	descWidth := m.width - 2 - nameMaxLen - 2 - keyMaxLen - 2
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
		Foreground(m.theme.Colors.Timestamp).
		Width(keyMaxLen).
		Render(cmd.Keybinding.Help().Key)

	// Join columns horizontally
	actionLine := lipgloss.JoinHorizontal(lipgloss.Top, nameStr, descStr, keyStr)

	return lipgloss.NewStyle().
		Background(backgroundColor).
		Padding(0, 1).
		Render(actionLine)
}
