package layout

const (
	InputBoxHeight = 3 // 2*border + margin
	InputBoxMargin = 1

	// Sidebars will occupy 10% of the terminal screen size
	channelSidebarRatio = 0.1
	userSidebarRatio    = 0.1

	sidebarPadding  = 2
	appPadding      = 2
	verticalPadding = 1
)

type Layout struct {
	ScreenWidth  int
	ScreenHeight int

	AppHeight int
	AppWidth  int

	SiderbarWidth int
}

// GenerateLayout computes the full layout based on the terminal screen size
func GenerateLayout(terminalWidth, terminalHeight int) Layout {
	appAvailableHeight := terminalHeight - verticalPadding
	if appAvailableHeight < 0 {
		appAvailableHeight = 0 // terminal too small
	}

	rawChannelsWidth := int(float64(terminalWidth) * channelSidebarRatio)
	rawUsersWidth := int(float64(terminalWidth) * userSidebarRatio)

	adjustedSidebarWidth := rawChannelsWidth - sidebarPadding
	adjustedAppWidth := terminalWidth - rawChannelsWidth - rawUsersWidth - appPadding

	if adjustedSidebarWidth < 0 {
		adjustedSidebarWidth = 0
	}
	if adjustedAppWidth < 0 {
		adjustedAppWidth = 0
	}

	totalHorizontalPadding := (sidebarPadding * 2) + appPadding

	return Layout{
		ScreenWidth:   terminalWidth - totalHorizontalPadding,
		ScreenHeight:  terminalHeight,
		AppHeight:     appAvailableHeight - verticalPadding,
		AppWidth:      adjustedAppWidth - 4,
		SiderbarWidth: adjustedSidebarWidth,
	}
}
