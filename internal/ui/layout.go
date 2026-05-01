package ui

import "charm.land/lipgloss/v2"

type layout struct {
	m *Model
}

func (l *layout) View() string {
	height := l.m.height
	if height < 0 {
		height = 0
	}

	channelsView := lipgloss.NewStyle().
		MaxWidth(channelsPanelMaxWidth).
		Render(l.m.channels.View(channelsPanelMaxWidth, height))

	buf := l.m.getActiveBuffer()
	usersView := lipgloss.NewStyle().
		MaxWidth(usersPanelMaxWidth).
		Render(buf.Users.View(usersPanelMaxWidth, height))

	chatView := lipgloss.NewStyle().
		Height(height).
		Render(buf.Chat.View())

	return lipgloss.JoinHorizontal(lipgloss.Top,
		channelsView,
		chatView,
		usersView,
	)
}
