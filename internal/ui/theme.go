package ui

import (
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/ui/components/palette"
	"github.com/dexchat/dex/internal/ui/styles"
)

var themeOptions = []palette.ThemeOption{
	{Name: config.ThemeRosePine, Label: "Rose Pine"},
	{Name: config.ThemeAyuDark, Label: "Ayu Dark"},
	{Name: config.ThemeDracula, Label: "Dracula"},
	{Name: config.ThemeGruvboxDark, Label: "Gruvbox Dark"},
	{Name: config.ThemeSolarizedLight, Label: "Solarized Light"},
}

func themeForName(name string) (styles.Theme, bool) {
	switch name {
	case config.ThemeRosePine:
		return styles.RosePineTheme(), true
	case config.ThemeAyuDark:
		return styles.AyuDarkTheme(), true
	case config.ThemeDracula:
		return styles.DraculaTheme(), true
	case config.ThemeGruvboxDark:
		return styles.GruvboxDarkTheme(), true
	case config.ThemeSolarizedLight:
		return styles.SolarizedLightTheme(), true
	default:
		return styles.RosePineTheme(), false
	}
}

func (m *Model) showThemePicker() {
	m.palette.ShowThemes(themeOptions, m.themeName)
}

func (m *Model) applyTheme(name string) {
	theme, ok := themeForName(name)
	if !ok {
		return
	}

	m.theme = theme
	m.themeName = name
	m.usernameColors = styles.NewUsernameColors(theme.Colors.Nicknames)
	m.channels = m.channels.SetTheme(theme)
	m.palette.SetTheme(theme)
	m.help.SetTheme(theme)
	for _, buffer := range m.buffers {
		buffer.Chat.SetTheme(theme, m.usernameColors)
		buffer.Users = buffer.Users.SetTheme(theme, m.usernameColors)
	}
}
