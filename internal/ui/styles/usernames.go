package styles

import (
	"math/rand"

	"github.com/charmbracelet/lipgloss"
)

type UsernameColors struct {
	colors  map[string]lipgloss.Color
	palette []lipgloss.Color
}

func NewUsernameColors(palette []lipgloss.Color) UsernameColors {
	return UsernameColors{
		colors:  map[string]lipgloss.Color{},
		palette: palette,
	}
}

func (u *UsernameColors) GetColor(username string) lipgloss.Color {
	if color, exists := u.colors[username]; exists {
		return color // Return cached color
	}
	// Assign new random color on first encounter
	color := u.palette[rand.Intn(len(u.palette))]
	u.colors[username] = color

	return color
}
