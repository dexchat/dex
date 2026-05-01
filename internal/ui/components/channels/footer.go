package channels

import (
	"charm.land/lipgloss/v2"
	"github.com/vaaleyard/dex/internal/ui/components/keybindings"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

func renderFooter(width int, theme styles.Theme) string {
	var footer string
	helpKeybinding := keybindings.DefaultKeyMap().TogglePalette
	quitKeybinding := keybindings.DefaultKeyMap().Quit

	defaultStyle := lipgloss.NewStyle().
		Background(theme.Colors.Base.Background)

	if keybindings.QuitHandlerIsWaitingForSecondPress() {
		styledKey := defaultStyle.
			Foreground(theme.Colors.Base.Foreground).
			Render(quitKeybinding.Help().Key)
		styledDesc := defaultStyle.
			Foreground(theme.Colors.Base.Error).
			Render(" again to exit")

		footer = styledKey + styledDesc
	} else {
		styledKey := defaultStyle.
			Foreground(theme.Colors.Base.Dimmed).
			Render(helpKeybinding.Help().Key)
		styledDesc := defaultStyle.
			Foreground(theme.Colors.Base.Subtle).
			Render(" for help")

		footer = styledKey + styledDesc
	}

	style := lipgloss.NewStyle().
		Background(theme.Colors.Base.Background).
		Width(width).
		Align(lipgloss.Center)

	return style.Render(footer)
}
