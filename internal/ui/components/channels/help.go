package channels

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type helpKeyMap struct {
	Help key.Binding
}

func renderHelp(width int, theme styles.Theme) string {
	helpKeys := helpKeyMap{
		Help: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "for help"),
		),
	}

	keyText := helpKeys.Help.Help().Key
	descText := helpKeys.Help.Help().Desc

	keyStyle := lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Foreground(theme.Colors.Timestamp)

	descStyle := lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Foreground(theme.Colors.StatusBg)

	styledKey := keyStyle.Render(keyText)
	styledDesc := descStyle.Render(descText)
	styledSpace := descStyle.Render(" ")
	s := styledKey + styledSpace + styledDesc

	style := lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Width(width).
		Align(lipgloss.Center)

	return style.Render(s)
}
