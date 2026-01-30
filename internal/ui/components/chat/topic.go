package chat

import "github.com/charmbracelet/lipgloss"

func (m *Model) renderTopic(width int) string {
	styledTopic := renderIRCFormattedMessage(m.topic, lipgloss.NewStyle())

	return lipgloss.NewStyle().
		Background(m.theme.Colors.LighterBackground).
		Foreground(m.theme.Colors.Accent).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		MarginBottom(1).
		Width(width + 2).
		Render(styledTopic)
}
