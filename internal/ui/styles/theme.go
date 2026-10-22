package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Colors Colors
	Styles Styles
}

type Color = color.Color

type Colors struct {
	Base struct {
		Background Color
		Foreground Color
		Surface    Color
		Border     Color
		Accent     Color
		Error      Color
		Dimmed     Color
		Subtle     Color
		OnAccent   Color
		Divider    Color
	}

	Sidebar struct {
		Server       Color
		Unread       Color
		Notification Color
		Mention      Color
		Selection    Color
	}

	Chat struct {
		Nickname         Color
		InactiveNickname Color
		Mention          Color
		Separator        Color
		Self             Color
		Connected        Color
		Disconnected     Color
		UserEvents       Color
		ServerMessage    Color
	}

	Palette struct {
		Border    Color
		Highlight Color
	}

	Nicknames []Color
}

type Styles struct {
	App           lipgloss.Style
	Sidebar       lipgloss.Style
	UnreadItem    lipgloss.Style
	NotifiedItem  lipgloss.Style
	MentionedItem lipgloss.Style
	ServerItem    lipgloss.Style
	ChatArea      lipgloss.Style
	InputField    lipgloss.Style
}

// LeftDividerBorder draws a vertical line on the left side only. The other
// sides are blank, so they show the border background.
func LeftDividerBorder() lipgloss.Border {
	border := lipgloss.HiddenBorder()
	border.Left = lipgloss.NormalBorder().Left
	return border
}

// RightDividerBorder draws a vertical line on the right side only. The other
// sides are blank, so they show the border background.
func RightDividerBorder() lipgloss.Border {
	border := lipgloss.HiddenBorder()
	border.Right = lipgloss.NormalBorder().Right
	return border
}

// HasBackground reports whether the theme paints its own background. Without
// one, the terminal's default background shows through.
func (t Theme) HasBackground() bool {
	_, isNoColor := t.Colors.Base.Background.(lipgloss.NoColor)
	return t.Colors.Base.Background != nil && !isNoColor
}
