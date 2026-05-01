package styles

import (
	"math/rand"
	"strings"
)

type UsernameColors struct {
	colors  map[string]Color
	palette []Color
}

func NewUsernameColors(palette []Color) UsernameColors {
	return UsernameColors{
		colors:  map[string]Color{},
		palette: palette,
	}
}

func (u *UsernameColors) GetColor(username string) Color {
	key := strings.ToLower(username)
	if color, exists := u.colors[key]; exists {
		return color
	}
	color := u.palette[rand.Intn(len(u.palette))]
	u.colors[key] = color

	return color
}
