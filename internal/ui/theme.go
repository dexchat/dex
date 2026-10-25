package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/ui/styles"
)

// themeDetectTimeout bounds how long theme = "auto" waits for the terminal
// to report its appearance before keeping the dark theme.
const themeDetectTimeout = time.Second

type themeDetectTimeoutMsg struct{}

func themeByName(name string) styles.Theme {
	switch name {
	case config.ThemeAyuDark:
		return styles.AyuDarkTheme()
	case config.ThemeDracula:
		return styles.DraculaTheme()
	case config.ThemeGruvboxDark:
		return styles.GruvboxDarkTheme()
	case config.ThemeSolarizedLight:
		return styles.SolarizedLightTheme()
	case config.ThemeNone:
		return styles.NoneTheme()
	default:
		return styles.RosePineTheme()
	}
}

// detectTheme asks the terminal for the system appearance, falling back to
// its background color for terminals that do not answer the light/dark
// report. The first answer wins; a later background answer and the timeout
// are ignored. It also enables light/dark reports, so the theme follows the
// system appearance until dex exits.
func (m *Model) detectTheme() tea.Cmd {
	if !m.themeDetecting {
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

// applySystemAppearance switches to theme_light or theme_dark when the
// terminal reports the system appearance, either answering detectTheme or
// announcing a change.
func (m *Model) applySystemAppearance(dark bool) {
	if !m.followSystemTheme {
		return
	}
	if !m.themeDetecting && dark == m.systemDark {
		return
	}
	m.themeDetecting = false
	m.systemDark = dark
	name := m.config.UI.ThemeLight
	if dark {
		name = m.config.UI.ThemeDark
	}
	m.setTheme(themeByName(name))
}

// applyTerminalBackground uses the terminal background as the appearance
// only while detecting. Afterwards the view may have set the background
// itself, so the answer would describe our theme rather than the system.
func (m *Model) applyTerminalBackground(dark bool) {
	if !m.themeDetecting {
		return
	}
	m.applySystemAppearance(dark)
}
