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

func TestPlainTextRemovesFormatting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "hello <world> & co", want: "hello <world> & co"},
		{name: "toggles", in: "\x02bold\x02 \x1ditalic\x1d \x1funder\x1f \x16rev\x16", want: "bold italic under rev"},
		{name: "numeric colors", in: "\x0304red\x03 \x0312,01blue on black\x0f done", want: "red blue on black done"},
		{name: "hex colors", in: "\x043FB950green\x04 \x03#FF0000red", want: "green red"},
		{name: "bare color resets", in: "\x03, not a color", want: ", not a color"},
		{name: "unicode", in: "\x02olá\x02 🎉", want: "olá 🎉"},
		{name: "only codes", in: "\x02\x0304\x0f", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlainText(tt.in); got != tt.want {
				t.Fatalf("PlainText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
