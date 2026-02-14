package styles

import (
	"math/rand"
	"strings"

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
	key := strings.ToLower(username)
	if color, exists := u.colors[key]; exists {
		return color
	}
	color := u.palette[rand.Intn(len(u.palette))]
	u.colors[key] = color

	return color
}
