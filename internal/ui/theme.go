package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/dexchat/dex/internal/ui/styles"
	"github.com/dexchat/dex/internal/ui/styles/themes"
)

// themeDetectTimeout bounds how long a light/dark theme waits for the
// terminal to report its appearance before keeping the dark theme.
const themeDetectTimeout = time.Second

// systemTheme tracks how a light/dark theme follows the system appearance.
type systemTheme int

const (
	// systemThemeOff means a single theme is configured.
	systemThemeOff systemTheme = iota
	// systemThemeDetecting waits for the first appearance report.
	systemThemeDetecting
	systemThemeLight
	systemThemeDark
)

type themeDetectTimeoutMsg struct{}

func themeByName(name string) styles.Theme {
	if theme, ok := themes.ByName(name); ok {
		return theme
	}
	theme, _ := themes.ByName(themes.Default)
	return theme
}

// detectTheme asks the terminal for the system appearance, falling back to
// its background color for terminals that do not answer the light/dark
// report. The first answer wins; a later background answer and the timeout
// are ignored. It also enables light/dark reports, so the theme follows the
// system appearance until dex exits.
func (m *Model) detectTheme() tea.Cmd {
	if m.systemTheme != systemThemeDetecting {
		return nil
	}
	return tea.Batch(
		tea.Sequence(
			tea.Raw(ansi.SetModeLightDark),
			tea.Raw(ansi.RequestLightDarkReport),
			tea.RequestBackgroundColor,
		),
		tea.Tick(themeDetectTimeout, func(time.Time) tea.Msg {
			return themeDetectTimeoutMsg{}
		}),
	)
}

// applySystemAppearance switches to the light or dark theme when the
// terminal reports the system appearance, either answering detectTheme or
// announcing a change.
func (m *Model) applySystemAppearance(dark bool) {
	if m.systemTheme == systemThemeOff {
		return
	}
	next, name := systemThemeLight, m.config.UI.SystemTheme.Light
	if dark {
		next, name = systemThemeDark, m.config.UI.SystemTheme.Dark
	}
	if m.systemTheme == next {
		return
	}
	m.systemTheme = next
	m.setTheme(themeByName(name))
}

// applyTerminalBackground uses the terminal background as the appearance
// only while detecting. Afterwards the view may have set the background
// itself, so the answer would describe our theme rather than the system.
func (m *Model) applyTerminalBackground(dark bool) {
	if m.systemTheme != systemThemeDetecting {
		return
	}
	m.applySystemAppearance(dark)
}
