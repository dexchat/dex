package styles

import (
	"image/color"
	"math"
	"testing"
)

func TestThemeTextContrast(t *testing.T) {
	themes := map[string]Theme{
		"ayu-dark":        AyuDarkTheme(),
		"gruvbox-dark":    GruvboxDarkTheme(),
		"rose-pine":       RosePineTheme(),
		"solarized-light": SolarizedLightTheme(),
	}

	for name, theme := range themes {
		t.Run(name, func(t *testing.T) {
			textColors := []Color{
				theme.Colors.Base.Foreground,
				theme.Colors.Base.Error,
				theme.Colors.Base.Dimmed,
				theme.Colors.Base.Subtle,
				theme.Colors.Sidebar.Server,
				theme.Colors.Sidebar.Unread,
				theme.Colors.Sidebar.Notification,
				theme.Colors.Sidebar.Mention,
				theme.Colors.Chat.Nickname,
				theme.Colors.Chat.InactiveNickname,
				theme.Colors.Chat.Mention,
				theme.Colors.Chat.Self,
				theme.Colors.Chat.Connected,
				theme.Colors.Chat.Disconnected,
				theme.Colors.Chat.UserEvents,
				theme.Colors.Chat.ServerMessage,
			}
			textColors = append(textColors, theme.Colors.Nicknames...)

			for _, textColor := range textColors {
				if ratio := contrastRatio(textColor, theme.Colors.Base.Background); ratio < 4.5 {
					t.Errorf("color %v has contrast %.2f:1 against background, want at least 4.5:1", textColor, ratio)
				}
			}

			if ratio := contrastRatio(theme.Colors.Base.Background, theme.Colors.Palette.Highlight); ratio < 4.5 {
				t.Errorf("palette selection has contrast %.2f:1, want at least 4.5:1", ratio)
			}
		})
	}
}

func TestSolarizedLightInteractiveContrast(t *testing.T) {
	theme := SolarizedLightTheme()
	assertContrast(t, theme.Colors.Base.Foreground, theme.Colors.Base.Surface)
	assertContrast(t, theme.Colors.Base.Accent, theme.Colors.Base.Surface)

	selectedText := []Color{
		theme.Colors.Base.Foreground,
		theme.Colors.Base.Dimmed,
		theme.Colors.Sidebar.Server,
		theme.Colors.Sidebar.Unread,
		theme.Colors.Sidebar.Notification,
		theme.Colors.Sidebar.Mention,
	}
	for _, textColor := range selectedText {
		assertContrast(t, textColor, theme.Colors.Sidebar.Selection)
	}
}

func assertContrast(t *testing.T, foreground, background color.Color) {
	t.Helper()
	if ratio := contrastRatio(foreground, background); ratio < 4.5 {
		t.Errorf("color %v has contrast %.2f:1 against %v, want at least 4.5:1", foreground, ratio, background)
	}
}

func contrastRatio(a, b color.Color) float64 {
	aLum := relativeLuminance(a)
	bLum := relativeLuminance(b)
	lighter := math.Max(aLum, bLum)
	darker := math.Min(aLum, bLum)
	return (lighter + 0.05) / (darker + 0.05)
}

func relativeLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*linearChannel(r) + 0.7152*linearChannel(g) + 0.0722*linearChannel(b)
}

func linearChannel(value uint32) float64 {
	channel := float64(value) / 65535
	if channel <= 0.04045 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}
