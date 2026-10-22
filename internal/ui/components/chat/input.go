package chat

import (
	"charm.land/lipgloss/v2"
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
	m.input.SetWidth(inputWidth)
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

	inputView := m.input.View()
	if len(m.pastedLines) > 0 {
		inputView = m.renderPastedLines()
	}
	inputView = nicknamePrefix + separator + inputView

	inputBox := m.theme.Styles.InputField.
		Margin(0, InputBoxMargin, 0, InputBoxMargin).
		MarginBackground(m.theme.Colors.Base.Background).
		Border(lipgloss.HiddenBorder()).
		BorderBackground(m.theme.Colors.Base.Surface).
		MarginTop(InputBoxPaddingTop).
		Render(inputView)

	return inputBox
}
