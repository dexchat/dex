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
		Foreground(m.usernameColors.GetColor(m.nickname)).
		Bold(true).
		Render(m.nickname)

	separator := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Chat.Separator).
		Background(m.theme.Colors.Base.Surface).
		Render(" | ")

	inputView := nicknamePrefix + separator + m.input.View()

	inputBox := m.theme.Styles.InputField.
		Margin(0, InputBoxMargin, 0, InputBoxMargin).
		MarginBackground(m.theme.Colors.Base.Background).
		Border(lipgloss.NormalBorder()).
		BorderBackground(m.theme.Colors.Base.Surface).
		BorderForeground(m.theme.Colors.Base.Surface).
		MarginTop(InputBoxPaddingTop).
		Render(inputView)

	return inputBox
}
