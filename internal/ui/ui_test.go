package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPaneForMouseWheelUsesPaneBounds(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want scrollPane
	}{
		{name: "channels", x: 0, want: scrollPaneChannels},
		{name: "last channel column", x: channelsPanelMaxWidth - 1, want: scrollPaneChannels},
		{name: "chat", x: channelsPanelMaxWidth, want: scrollPaneChat},
		{name: "users", x: 100 - usersPanelMaxWidth, want: scrollPaneUsers},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paneForMouseWheel(tea.MouseWheelMsg{
				X:      tt.x,
				Button: tea.MouseWheelDown,
			}, 100)
			if got != tt.want {
				t.Fatalf("paneForMouseWheel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaneForMouseWheelIgnoresNonWheelMouse(t *testing.T) {
	got := paneForMouseWheel(tea.MouseClickMsg{
		X:      0,
		Button: tea.MouseLeft,
	}, 100)
	if got != scrollPaneNone {
		t.Fatalf("paneForMouseWheel() = %v, want %v", got, scrollPaneNone)
	}
}
