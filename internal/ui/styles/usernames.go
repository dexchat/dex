package styles

import (
	"math/rand"
	"strings"
)

// UsernameColors assigns each nickname a slot in the theme's nickname
// palette. Copies share the assignments, so every component sees the same
// color for a nickname.
type UsernameColors struct {
	slots   map[string]int
	palette []Color
}

func NewUsernameColors(palette []Color) UsernameColors {
	return UsernameColors{
		slots:   map[string]int{},
		palette: palette,
	}
}

// WithPalette returns a copy that reads colors from palette while keeping
// the existing assignments, so a theme switch moves each nickname to the
// same slot of the new palette instead of picking a new random one.
func (u UsernameColors) WithPalette(palette []Color) UsernameColors {
	u.palette = palette
	return u
}

func (u *UsernameColors) GetColor(username string) Color {
	key := strings.ToLower(username)
	slot, exists := u.slots[key]
	if !exists {
		slot = rand.Intn(len(u.palette))
		u.slots[key] = slot
	}

	return u.palette[slot%len(u.palette)]
}
