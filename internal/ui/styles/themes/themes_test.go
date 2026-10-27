package themes

import (
	"image/color"
	"math"
	"testing"

	"github.com/dexchat/dex/internal/ui/styles"
)

func TestThemeTextContrast(t *testing.T) {
	forEachColorTheme(t, func(t *testing.T, theme styles.Theme) {
		textColors := []styles.Color{
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

		if ratio := contrastRatio(theme.Colors.Base.OnAccent, theme.Colors.Palette.Highlight); ratio < 4.5 {
			t.Errorf("palette selection has contrast %.2f:1, want at least 4.5:1", ratio)
		}
	})
}

func TestThemeInteractiveContrast(t *testing.T) {
	forEachColorTheme(t, func(t *testing.T, theme styles.Theme) {
		assertContrast(t, theme.Colors.Base.Foreground, theme.Colors.Base.Surface)
		assertContrast(t, theme.Colors.Base.Accent, theme.Colors.Base.Surface)

		selectedText := []styles.Color{
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
	})
}

// forEachColorTheme runs test for every built-in theme that paints its own
// background. Themes that use the terminal's palette cannot be measured.
func forEachColorTheme(t *testing.T, test func(t *testing.T, theme styles.Theme)) {
	t.Helper()
	for _, name := range Names() {
		theme, _ := ByName(name)
		if !theme.HasBackground() {
			continue
		}
		t.Run(name, func(t *testing.T) {
			test(t, theme)
		})
	}
}

func TestThemeByNameFindsEveryTheme(t *testing.T) {
	for _, name := range Names() {
		if _, ok := ByName(name); !ok {
			t.Errorf("ByName(%q) not found", name)
		}
	}
	if _, ok := ByName(Default); !ok {
		t.Errorf("default theme %q is not registered", Default)
	}
	if _, ok := ByName("rose-pink"); ok {
		t.Error("ByName(\"rose-pink\") found an unknown theme")
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
