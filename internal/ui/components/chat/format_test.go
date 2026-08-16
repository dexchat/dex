package chat

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestParseIRCFormatHexColor(t *testing.T) {
	segments := parseIRCFormat("\x043FB950[Approved]", lipgloss.NewStyle())

	if len(segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(segments))
	}
	if segments[0].text != "[Approved]" {
		t.Fatalf("text = %q, want %q", segments[0].text, "[Approved]")
	}

	got := segments[0].style.GetForeground()
	want := lipgloss.Color("#3FB950")
	if !sameColor(got, want) {
		t.Fatalf("foreground = %#v, want %#v", got, want)
	}
}

func TestParseIRCFormatNumericColor(t *testing.T) {
	segments := parseIRCFormat("\x034[Approved]", lipgloss.NewStyle())

	if len(segments) != 1 || segments[0].text != "[Approved]" {
		t.Fatalf("segments = %#v, want one segment containing [Approved]", segments)
	}
	if !sameColor(segments[0].style.GetForeground(), ircColors[4]) {
		t.Fatalf("foreground = %#v, want %#v", segments[0].style.GetForeground(), ircColors[4])
	}
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
