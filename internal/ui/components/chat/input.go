package chat

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	InputBoxMargin     = 1
	InputBoxPaddingTop = 1
)

func (m *Model) setInputWidth(width int) {
	prefixWidth := len(m.nickname) + len(" | ")
	inputWidth := width - prefixWidth - InputBoxMargin*2 - 1

	if inputWidth < 12 { // minimum width to not panic
		inputWidth = 12
	}
	m.input.Width = inputWidth
}

func (m *Model) renderInputBox() string {
	nicknamePrefix := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Accent).
		Bold(true).
		Render(m.nickname)

	separator := lipgloss.NewStyle().
		Foreground(m.theme.Colors.BorderColor).
		Background(m.theme.Colors.LighterBackground).
		Render(" | ")

	inputView := nicknamePrefix + separator + m.input.View()

	inputBox := m.theme.Styles.InputField.
		Margin(0, InputBoxMargin, 0, InputBoxMargin).
		MarginBackground(m.theme.Colors.Background).
		Border(lipgloss.NormalBorder()).
		BorderBackground(m.theme.Colors.LighterBackground).
		BorderForeground(m.theme.Colors.LighterBackground).
		MarginTop(InputBoxPaddingTop).
		Render(inputView)

	return inputBox
}
