package ui

import "charm.land/lipgloss/v2"

func overlayCenter(foreground, background string) string {
	if foreground == "" {
		return background
	}
	if background == "" {
		return foreground
	}

	bgWidth, bgHeight := lipgloss.Size(background)
	if bgWidth <= 0 || bgHeight <= 0 {
		return foreground
	}

	fgWidth, fgHeight := lipgloss.Size(foreground)
	x := (bgWidth - fgWidth) / 2
	y := (bgHeight - fgHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	backgroundLayer := lipgloss.NewLayer(background)
	foregroundLayer := lipgloss.NewLayer(foreground).X(x).Y(y).Z(1)
	compositor := lipgloss.NewCompositor(backgroundLayer, foregroundLayer)

	return lipgloss.NewCanvas(bgWidth, bgHeight).
		Compose(compositor).
		Render()
}
