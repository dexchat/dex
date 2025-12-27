package chat

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/input"
	"github.com/vaaleyard/dex/internal/ui/layout"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

type Message struct {
	Nick string
	Text string
	Time string
}
type Model struct {
	messages []string
	viewport viewport.Model
	input    input.Model
	topic    string

	theme styles.Theme
}

func New(theme styles.Theme) Model {
	m := Model{
		messages: []string{
			"@idlebot gilfoyle found a pair of Nikes! This wondrous godsend has accelerated them 0 days, 03:49:14 towards level 97.",
			"@idlebot gilfoyle reaches next level in 3 days, 00:35:36.",
			"@idlebot dinesh [1190/1453] has come upon jared [821/1313] and taken them in combat! 12 days, 12:44:38 is removed from dinesh's clock.",
			"@idlebot dinesh reaches next level in 44 days, 10:16:26.",
			"@idlebot monica [651/822] has challenged richard [514/1140] in combat and won! 1 day, 21:22:17 is removed from monica's clock.",
			"@idlebot monica reaches next level in 5 days, 09:08:03.",
			"@idlebot In the fierce battle, richard dropped their level 75 helm! monica picks it up, tossing their old level 72 helm to richard.",
			"@idlebot gilfoyle [563/985] has come upon dinesh [949/1419] and been defeated in combat! 6 days, 03:38:17 is added to gilfoyle's clock.",
			"@idlebot gilfoyle reaches next level in 53 days, 11:19:01.",
			"@idlebot jared ate a poisonous fruit. This terrible calamity has slowed them 0 days, 15:03:16 from level 105.",
			"@idlebot jared reaches next level in 8 days, 11:14:18.",
			"@idlebot monica invented the wheel! This wondrous godsend has accelerated them 2 days, 19:58:55 towards level 95.",
			"@idlebot monica reaches next level in 20 days, 18:32:07.",
			"@idlebot richard was set on fire. This terrible calamity has slowed them 5 days, 12:16:31 from level 69.",
			"@idlebot richard reaches next level in 51 days, 10:34:14.",
			"@idlebot gilfoyle got a kiss from dinesh! This wondrous godsend has accelerated them 0 days, 08:33:29 towards level 97.",
			"@idlebot gilfoyle reaches next level in 2 days, 14:45:34.",
			"@idlebot jared, the Gangsta, has attained level 71! Next level in 62 days, 04:22:00.",
			"@idlebot jared [226/588] has challenged monica [412/993] in combat and lost! 4 days, 23:23:21 is added to jared's clock.",
			"@idlebot jared reaches next level in 67 days, 03:45:21.",
			"@idlebot richard had to fix some Whitespace code. This terrible calamity has slowed them 0 days, 00:21:38 from level 31.",
			"@idlebot richard reaches next level in 0 days, 03:38:20.",
			"@idlebot gilfoyle [384/898] has come upon dinesh [35/784] and taken them in combat! 8 days, 12:13:23 is removed from gilfoyle's clock.",
			"@idlebot gilfoyle reaches next level in 28 days, 11:42:16.",
			"@idlebot gilfoyle has dealt dinesh a Critical Strike! 0 days, 00:08:48 is added to dinesh's clock.",
			"@idlebot dinesh reaches next level in 0 days, 01:07:29.",
			"@idlebot jared, the \"Cook\", has attained level 100! Next level in 91 days, 04:22:00.",
			"@idlebot jared [1121/1241] has challenged monica [829/1222] in combat and won! 20 days, 23:19:27 is removed from jared's clock.",
			"@idlebot jared reaches next level in 70 days, 05:02:33.",
			"@idlebot richard [500/775] has come upon gilfoyle [1040/1136] and been defeated in combat! 3 days, 16:36:24 is added to richard's clock.",
			"@idlebot richard reaches next level in 44 days, 17:07:40.",
			"@idlebot dinesh found a pair of Nikes! This wondrous godsend has accelerated them 5 days, 15:04:26 towards level 95.",
			"@idlebot dinesh reaches next level in 64 days, 17:21:04.",
			"@idlebot jared encounters simple and bows humbly.",
			"@idlebot jared [206/586] has challenged monica [740/1260] in combat and lost! 5 days, 09:34:55 is added to jared's clock.",
			"@idlebot jared reaches next level in 50 days, 09:25:58.",
			"@idlebot richard gained a sixth sense! This wondrous godsend has accelerated them 1 day, 15:01:39 towards level 100.",
			"@idlebot richard reaches next level in 14 days, 15:14:53.",
			"@idlebot gilfoyle had to fix some Whitespace code. This terrible calamity has slowed them 5 days, 23:08:25 from level 66.",
			"@idlebot gilfoyle reaches next level in 60 days, 04:24:57.",
			"@idlebot dinesh is forsaken by their evil god. 1 day, 02:18:21 is added to their clock.",
			"@idlebot dinesh reaches next level in 37 days, 15:10:34.",
			"@idlebot jared fell, chipping the stone in their amulet! jared's amulet loses 10% of its effectiveness.",
			"@idlebot monica [745/939] has come upon richard [1203/1223] and been defeated in combat! 2 days, 01:04:53 is added to monica's clock.",
			"@idlebot monica reaches next level in 16 days, 15:39:46.",
			"@idlebot richard, the Paladin, has attained level 94! Next level in 85 days, 04:22:00.",
			"@idlebot richard [108/784] has challenged gilfoyle [1226/1297] in combat and lost! 11 days, 01:46:03 is added to richard's clock.",
			"@idlebot richard reaches next level in 96 days, 06:08:03.",
			"@idlebot dinesh reinforced their shield with a dragon's scales! dinesh's shield gains 10% effectiveness.",
			"@idlebot jared [876/1004] has come upon monica [1065/1279] and been defeated in combat! 3 days, 21:25:53 is added to jared's clock.",
			"@idlebot jared reaches next level in 39 days, 06:48:30.",
		},
		topic:    "Welcome to the Libera IdleRPG game.  Discussion in #idlerpg-discuss | Website: https://idlerpg.lolhosting.net | Please read: https://idlerpg.lolhosting.net#conduct",
		theme:    theme,
		input:    input.New(theme),
		viewport: viewport.New(0, 0),
	}
	m.viewport.SetContent(strings.Join(m.messages, "\n"))

	return m
}

func (m Model) Init() tea.Cmd {
	return m.input.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmd := m.input.Update(msg)
	return m, cmd
}

func (m Model) View(width, height int) string {
	topicContent := lipgloss.NewStyle().
		Background(m.theme.Colors.LighterBackground).
		Foreground(m.theme.Colors.Accent).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Width(width + 4). // 2 for both sides
		Render(m.topic)
	topicView := lipgloss.NewStyle().
		PaddingTop(1).
		Render(topicContent)

	topicLines := strings.Count(topicContent, "\n") + 1

	// calculate topic line count dynamically to correct render it
	m.viewport.Width = width
	m.viewport.Height = height - layout.InputBoxHeight - topicLines - 1

	inputView := m.input.View(width)

	paddedInputView := lipgloss.NewStyle().
		PaddingTop(1).
		Render(inputView)

	paddedViewport := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Render(m.viewport.View())

	combinedView := lipgloss.JoinVertical(
		lipgloss.Left,
		topicView,
		paddedViewport,
		paddedInputView,
	)

	return m.theme.Styles.ChatArea.
		Render(combinedView)
}
