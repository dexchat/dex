package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlushSnapshotReplacesHistoryWithoutTemporaryFiles(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	entries := []LogEntry{{
		ServerTime: 1,
		Username:   "alice",
		Text:       "hello",
	}}

	if err := FlushSnapshot("libera", "#go", entries); err != nil {
		t.Fatalf("FlushSnapshot() error = %v", err)
	}

	loaded := Load("libera", "#go").Entries()
	if len(loaded) != 1 || loaded[0].Text != "hello" {
		t.Fatalf("Load() = %#v, want one hello entry", loaded)
	}

	files, err := os.ReadDir(filepath.Dir(logFullPath("libera", "#go")))
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".history-") {
			t.Fatalf("temporary history file was left behind: %s", file.Name())
		}
	}
}
