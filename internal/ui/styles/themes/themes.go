package themes

import "github.com/dexchat/dex/internal/ui/styles"

// Default is the theme used when the configuration names none.
const Default = "rose-pine"

// themes lists every built-in theme by its configuration name. Adding a
// theme means writing its constructor and adding it here.
var themes = []struct {
	name string
	new  func() styles.Theme
}{
	{name: "rose-pine", new: RosePine},
	{name: "rose-pine-dawn", new: RosePineDawn},
	{name: "ayu-dark", new: AyuDark},
	{name: "dracula", new: Dracula},
	{name: "gruvbox-dark", new: GruvboxDark},
	{name: "solarized-light", new: SolarizedLight},
	{name: "none", new: None},
}

// ByName returns the built-in theme with the given configuration name.
func ByName(name string) (styles.Theme, bool) {
	for _, theme := range themes {
		if theme.name == name {
			return theme.new(), true
		}
	}
	return styles.Theme{}, false
}

// Names returns the configuration names of the built-in themes.
func Names() []string {
	names := make([]string, len(themes))
	for i, theme := range themes {
		names[i] = theme.name
	}
	return names
}
