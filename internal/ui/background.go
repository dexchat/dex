package ui

import "github.com/charmbracelet/lipgloss"

type background struct {
	m *model
}

func (b *background) View() string {
	adjustedHeight := b.m.Height - topPadding
	if adjustedHeight < 0 {
		adjustedHeight = 0
	}

	channelsView := lipgloss.NewStyle().
		MaxWidth(channelsPanelMaxWidth).
		PaddingTop(topPadding).
		Render(b.m.channels.View(channelsPanelMaxWidth, adjustedHeight))

	usersView := lipgloss.NewStyle().
		MaxWidth(usersPanelMaxWidth).
		PaddingTop(topPadding).
		Render(b.m.users.View(usersPanelMaxWidth, adjustedHeight))

	chatView := lipgloss.NewStyle().
		PaddingTop(topPadding).
		Height(adjustedHeight).
		Render(b.m.chat.View())

	return lipgloss.JoinHorizontal(lipgloss.Top,
		channelsView,
		chatView,
		usersView,
	)
}
