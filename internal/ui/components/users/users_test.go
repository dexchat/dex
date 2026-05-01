package users

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

func TestMouseWheelScrollsUserList(t *testing.T) {
	m := newScrollableUsersModel()

	var cmd tea.Cmd
	m, cmd = m.Update(tea.MouseWheelMsg{
		Button: tea.MouseWheelDown,
	})
	if cmd != nil {
		_ = cmd()
	}

	if m.viewport.YOffset() == 0 {
		t.Fatal("expected mouse wheel down to scroll the user list")
	}
}

func TestUserPaneRendersRequestedHeight(t *testing.T) {
	m := newScrollableUsersModel()

	lines := strings.Split(m.View(20, 10), "\n")
	if got, want := len(lines), 10; got != want {
		t.Fatalf("expected rendered user pane height %d, got %d", want, got)
	}
}

func newScrollableUsersModel() Model {
	theme := styles.RosePineTheme()
	m := New(theme, styles.NewUsernameColors(theme.Colors.Nicknames))
	m = m.SetSize(20, 10)

	var userNames []string
	for i := 0; i < 30; i++ {
		userNames = append(userNames, fmt.Sprintf("user-%02d", i))
	}

	m, _ = m.Update(UserListMsg(userNames))
	return m
}
