package layout

var (
	TabRowHeight   = 1
	InputBoxHeight = 1

	// Sidebars will occupy 10% of the terminal screen size
	ChannelSidebarRatio = 0.1
	UserSidebarRatio    = 0.1

	TabRowPadding         = 0
	InputBoxPadding       = 1
	ChannelSidebarPadding = 2
	UserSidebarPadding    = 2
	MainContentPadding    = 2
	VerticalPadding       = 2
)

// Layout holds computed widths/heights for all panes
type Layout struct {
	ScreenWidth  int
	ScreenHeight int

	TabRowHeight      int
	InputBoxHeight    int
	MainContentHeight int

	ChannelSidebarWidth int
	MainContentWidth    int
	UserSidebarWidth    int
	InputBoxWidth       int
}

// GenerateLayout computes the full layout based on the terminal screen size
func GenerateLayout(terminalWidth, terminalHeight int, isTabHidden bool) Layout {
	if isTabHidden {
		TabRowHeight = 0
	}

	mainContentAvailableHeight := terminalHeight - TabRowHeight - InputBoxHeight - VerticalPadding
	if mainContentAvailableHeight < 0 {
		mainContentAvailableHeight = 0 // terminal too small
	}

	rawChannelsWidth := int(float64(terminalWidth) * ChannelSidebarRatio)
	rawUsersWidth := int(float64(terminalWidth) * UserSidebarRatio)

	// Calculate widths to account for padding
	adjustedChannelSidebarWidth := rawChannelsWidth - ChannelSidebarPadding
	adjustedUserSidebarWidth := rawUsersWidth - UserSidebarPadding
	adjustedMainContentWidth := terminalWidth - rawChannelsWidth - rawUsersWidth - MainContentPadding
	adjustedInputBoxWidth := terminalWidth - (InputBoxPadding * 2)

	if adjustedChannelSidebarWidth < 0 {
		adjustedChannelSidebarWidth = 0
	}
	if adjustedUserSidebarWidth < 0 {
		adjustedUserSidebarWidth = 0
	}
	if adjustedMainContentWidth < 0 {
		adjustedMainContentWidth = 0
	}

	totalHorizontalPadding := ChannelSidebarPadding + UserSidebarPadding + MainContentPadding

	return Layout{
		ScreenWidth:         terminalWidth - totalHorizontalPadding,
		ScreenHeight:        terminalHeight,
		TabRowHeight:        TabRowHeight,
		InputBoxHeight:      InputBoxHeight,
		MainContentHeight:   mainContentAvailableHeight - VerticalPadding,
		ChannelSidebarWidth: adjustedChannelSidebarWidth,
		MainContentWidth:    adjustedMainContentWidth,
		UserSidebarWidth:    adjustedUserSidebarWidth,
		InputBoxWidth:       adjustedInputBoxWidth,
	}
}
