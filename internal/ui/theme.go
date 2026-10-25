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
// report. The first answer wins; later ones and the timeout are ignored.
func (m *Model) detectTheme() tea.Cmd {
	if !m.themeDetecting {
		return nil
	}
	return tea.Batch(
		tea.Sequence(tea.Raw(ansi.RequestLightDarkReport), tea.RequestBackgroundColor),
		tea.Tick(themeDetectTimeout, func(time.Time) tea.Msg {
			return themeDetectTimeoutMsg{}
		}),
	)
}

func (m *Model) applyDetectedTheme(dark bool) {
	if !m.themeDetecting {
		return
	}
	m.themeDetecting = false
	name := m.config.UI.ThemeLight
	if dark {
		name = m.config.UI.ThemeDark
	}
	m.setTheme(themeByName(name))
}
