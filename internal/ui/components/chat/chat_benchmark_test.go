package chat

import (
	"fmt"
	"testing"
	"time"

	"github.com/dexchat/dex/internal/ui/styles"
)

func BenchmarkFlushQueue(b *testing.B) {
	for _, historySize := range []int{100, 1_000, 10_000} {
		for _, batchSize := range []int{1, 50} {
			b.Run(fmt.Sprintf("history=%d/batch=%d", historySize, batchSize), func(b *testing.B) {
				theme := styles.RosePineTheme()
				base := benchmarkMessages(historySize, time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC))
				batch := benchmarkMessages(batchSize, base[len(base)-1].Timestamp.Add(time.Minute))

				b.ReportAllocs()
				for range b.N {
					b.StopTimer()
					m := newBenchmarkChat(theme, base)
					for _, msg := range batch {
						m.QueueMessage(msg)
					}
					b.StartTimer()

					m.FlushQueue()
				}
			})
		}
	}
}

func BenchmarkRefreshContent(b *testing.B) {
	for _, historySize := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("history=%d", historySize), func(b *testing.B) {
			theme := styles.RosePineTheme()
			messages := benchmarkMessages(historySize, time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC))
			m := newBenchmarkChat(theme, messages)

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				m.RefreshContent()
			}
		})
	}
}

func BenchmarkSetSizeWidthChange(b *testing.B) {
	for _, historySize := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("history=%d", historySize), func(b *testing.B) {
			theme := styles.RosePineTheme()
			messages := benchmarkMessages(historySize, time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC))

			b.ReportAllocs()
			for range b.N {
				b.StopTimer()
				m := newBenchmarkChat(theme, messages)
				b.StartTimer()

				m.SetSize(101, 24)
			}
		})
	}
}

func newBenchmarkChat(theme styles.Theme, messages []Message) Model {
	m := New(theme, styles.NewUsernameColors(theme.Colors.Nicknames))
	m.SetNickname("dexuser")
	m.SetSize(100, 24)
	m.ReplaceMessages(messages)
	m.FlushQueue()
	return m
}

func benchmarkMessages(count int, start time.Time) []Message {
	messages := make([]Message, count)
	for i := range messages {
		messages[i] = Message{
			Timestamp: start.Add(time.Duration(i) * time.Minute),
			Username:  fmt.Sprintf("user-%d", i%20),
			Text:      fmt.Sprintf("\x02message %05d\x0f for dexuser: a formatted IRC chat history benchmark entry", i),
		}
	}
	return messages
}
