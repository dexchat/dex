package styles

import (
	"slices"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestWithPaletteKeepsNicknameSlots(t *testing.T) {
	first := []Color{lipgloss.Color("#000001"), lipgloss.Color("#000002"), lipgloss.Color("#000003")}
	second := []Color{lipgloss.Color("#100001"), lipgloss.Color("#100002"), lipgloss.Color("#100003")}
	colors := NewUsernameColors(first)
	before := colors.GetColor("Alice")
	slot := slices.Index(first, before)

	switched := colors.WithPalette(second)

	if got, want := switched.GetColor("alice"), second[slot]; got != want {
		t.Fatalf("color after palette switch = %v, want %v", got, want)
	}
	if got := colors.GetColor("alice"); got != before {
		t.Fatalf("original palette color = %v, want %v", got, before)
	}
}
