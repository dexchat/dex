package styles

import "github.com/charmbracelet/lipgloss"

type Tab struct {
	ActiveTab   lipgloss.Style
	InactiveTab lipgloss.Style
	TabBar      lipgloss.Style
}

func NewTabStyle() Tab {
	tabStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, false, false, true).
		BorderForeground(SecondaryColor).
		Padding(0, 3).
		Background(lipgloss.Color("240")).
		Foreground(TextColor)

	inactiveText := lipgloss.Color("#AAAAAA")

	activeTab := tabStyle.
		BorderForeground(PrimaryColor).
		Foreground(PrimaryColor)

	inactiveTab := lipgloss.NewStyle().
		Foreground(inactiveText).
		Padding(0, 2)

	tabBar := lipgloss.NewStyle()

	return Tab{
		ActiveTab:   activeTab,
		InactiveTab: inactiveTab,
		TabBar:      tabBar,
	}
}
