package channels

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vaaleyard/dex/internal/config"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

func TestMouseWheelScrollsListWithoutChangingSelection(t *testing.T) {
	m := newScrollableChannelsModel()
	beforeSelection := m.Selected()

	var cmd tea.Cmd
	m, cmd = m.Update(tea.MouseWheelMsg{
		Button: tea.MouseWheelDown,
	})
	if cmd != nil {
		_ = cmd()
	}

	if m.viewport.YOffset() == 0 {
		t.Fatal("expected mouse wheel down to scroll the channels list")
	}

	afterSelection := m.Selected()
	if afterSelection != beforeSelection {
		t.Fatalf("expected scrolling not to change selected channel, before=%+v after=%+v", beforeSelection, afterSelection)
	}
}

func TestChannelRowsRenderFullViewportWidth(t *testing.T) {
	m := newScrollableChannelsModel()

	for _, line := range strings.Split(m.viewport.GetContent(), "\n") {
		if line == "" {
			continue
		}
		if got, want := lipgloss.Width(line), m.viewport.Width(); got != want {
			t.Fatalf("expected channel row width %d, got %d for %q", want, got, line)
		}
	}
}

func TestChannelListKeepsOneRowBottomPadding(t *testing.T) {
	m := newScrollableChannelsModel()

	if got, want := m.viewport.VisibleLineCount(), m.viewport.Height()-1; got != want {
		t.Fatalf("expected channel viewport to reserve one bottom padding row, visible lines=%d want=%d", got, want)
	}
}

func TestChannelPaneRendersRequestedHeight(t *testing.T) {
	m := newScrollableChannelsModel()

	lines := strings.Split(m.View(25, 10), "\n")
	if got, want := len(lines), 10; got != want {
		t.Fatalf("expected rendered channel pane height %d, got %d", want, got)
	}
}

func newScrollableChannelsModel() Model {
	theme := styles.RosePineTheme()
	var channelNames []string
	for i := 0; i < 30; i++ {
		channelNames = append(channelNames, fmt.Sprintf("#channel-%02d", i))
	}

	m := New(theme, []*config.Server{
		{
			Name:     "libera",
			Channels: channelNames,
		},
	})
	return m.SetSize(25, 10)
}
