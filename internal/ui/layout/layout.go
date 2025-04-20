package layout

const (
	InputBoxHeight = 3 // top/bottom border + text input (1 line)
	InputBoxMargin = 1

	// Sidebars will occupy 10% of the terminal screen size
	channelSidebarRatio = 0.1
	userSidebarRatio    = 0.1

	sidebarsPadding = 2
	appPadding      = 2
	verticalPadding = 1
)

type Layout struct {
	ScreenWidth  int
	ScreenHeight int

	AppHeight     int
	ChatWidth     int
	SiderbarWidth int
}

// GenerateLayout computes the full layout based on the terminal screen size
func GenerateLayout(terminalWidth, terminalHeight int) Layout {
	appAvailableHeight := terminalHeight - verticalPadding
	if appAvailableHeight < 0 {
		appAvailableHeight = 0 // terminal too small, TODO: prevent app to open in these checks
	}

	rawChannelsWidth := int(float64(terminalWidth) * channelSidebarRatio)
	rawUsersWidth := int(float64(terminalWidth) * userSidebarRatio)

	adjustedSidebarWidth := rawChannelsWidth - sidebarsPadding
	adjustedChatWidth := terminalWidth - rawChannelsWidth - rawUsersWidth - appPadding

	if adjustedSidebarWidth < 0 {
		adjustedSidebarWidth = 0
	}
	if adjustedChatWidth < 0 {
		adjustedChatWidth = 0
	}

	totalHorizontalPadding := (sidebarsPadding * 2) + appPadding

	// chat left border (1) + chat right border (1) +
	// right sidebar left border (1) + left sidebar right border (1)
	horizontalBorderSum := 4

	return Layout{
		ScreenWidth:   terminalWidth - totalHorizontalPadding,
		ScreenHeight:  terminalHeight,
		AppHeight:     appAvailableHeight - verticalPadding,
		ChatWidth:     adjustedChatWidth - horizontalBorderSum,
		SiderbarWidth: adjustedSidebarWidth,
	}
}
