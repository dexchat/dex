package ui

import "github.com/charmbracelet/lipgloss"

type background struct {
	m *Model
}

func (b *background) View() string {
	adjustedHeight := b.m.height - topPadding
	if adjustedHeight < 0 {
		adjustedHeight = 0
	}

	channelsView := lipgloss.NewStyle().
		MaxWidth(channelsPanelMaxWidth).
		PaddingTop(topPadding).
		Render(b.m.channels.View(channelsPanelMaxWidth, adjustedHeight))

	buf := b.m.getActiveBuffer()
	usersView := lipgloss.NewStyle().
		MaxWidth(usersPanelMaxWidth).
		PaddingTop(topPadding).
		Render(buf.Users.View(usersPanelMaxWidth, adjustedHeight))

	chatView := lipgloss.NewStyle().
		PaddingTop(topPadding).
		Height(adjustedHeight).
		Render(buf.Chat.View())

	return lipgloss.JoinHorizontal(lipgloss.Top,
		channelsView,
		chatView,
		usersView,
	)
}
