package channels

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

func renderFooter(width int, theme styles.Theme) string {
	var footer string
	helpKeybinding := keybindings.DefaultKeyMap().TogglePalette
	quitKeybinding := keybindings.DefaultKeyMap().Quit

	defaultStyle := lipgloss.NewStyle().
		Background(theme.Colors.Background)

	if keybindings.QuitHandlerIsWaitingForSecondPress() {
		styledKey := defaultStyle.
			Foreground(theme.Colors.Text).
			Render(quitKeybinding.Help().Key)
		styledDesc := defaultStyle.
			Foreground(theme.Colors.ErrorMsg).
			Render(" again to exit")

		footer = styledKey + styledDesc
	} else {

		styledKey := defaultStyle.
			Foreground(theme.Colors.Timestamp).
			Render(helpKeybinding.Help().Key)
		styledDesc := defaultStyle.
			Foreground(theme.Colors.StatusBg).
			Render(" for help")

		footer = styledKey + styledDesc
	}

	style := lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Width(width).
		Align(lipgloss.Center)

	return style.Render(footer)
}
