package chat

import (
	"github.com/vaaleyard/dex/internal/ui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Message struct {
	Nick string
	Text string
	Time string
}
type Model struct {
	Messages []Message
}

func New() Model {
	return Model{
		Messages: []Message{
			{"amora", "hey everyone", "10:01"},
			{"jared", "sup!", "10:02"},
			{"gilfoyle", "what", "10:03"},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(width int, height int) string {
	var lines []string

	for _, msg := range m.Messages {
		nick := styles.UsernameStyle(msg.Nick)
		text := styles.UserMsgStyle.Render(msg.Text)
		timestamp := styles.TimestampStyle("[" + msg.Time + "]")

		lines = append(lines, timestamp+" "+nick+" "+text)
	}

	body := strings.Join(lines, "\n")

	return styles.ChatViewStyle.Width(width).Height(height).Render(body)
}
