package history

import (
	"compress/gzip"
	"errors"
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

	loaded := mustLoad(t, "libera", "#go").Entries()
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

func TestActionSurvivesPersistenceAndIsNotADuplicateOfPlainText(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	action := LogEntry{ServerTime: 1, Username: "alice", Text: "waves", Action: true}
	plain := LogEntry{ServerTime: 1, Username: "alice", Text: "waves"}

	log := NewLog()
	log.Insert(action)
	if log.IsDuplicate(plain) {
		t.Fatal("a plain message should not duplicate an action with the same text")
	}

	if err := FlushSnapshot("libera", "#go", []LogEntry{action}); err != nil {
		t.Fatalf("FlushSnapshot() error = %v", err)
	}
	loaded := mustLoad(t, "libera", "#go").Entries()
	if len(loaded) != 1 || !loaded[0].Action || loaded[0].Text != "waves" {
		t.Fatalf("Load() = %#v, want the action entry", loaded)
	}
}

func mustLoad(t *testing.T, server, buffer string) *Log {
	t.Helper()
	log, err := Load(server, buffer)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return log
}

// writeRawHistory writes gzip-compressed content directly to the history
// path, bypassing the encoder so tests can store damaged data.
func writeRawHistory(t *testing.T, server, buffer, content string) string {
	t.Helper()
	path := logFullPath(server, buffer)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	gz := gzip.NewWriter(file)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatalf("gzip Write() error = %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip Close() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return path
}

func TestLoadMissingFileReturnsEmptyLog(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	if got := len(mustLoad(t, "libera", "#go").Entries()); got != 0 {
		t.Fatalf("Load() entries = %d, want 0", got)
	}
}

func TestLoadCorruptFileReturnsReadableEntries(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeRawHistory(t, "libera", "#go", `{"server_time":1,"username":"alice","text":"hello"}`+"\n{broken")

	log, err := Load("libera", "#go")
	if !errors.Is(err, ErrCorrupt) {
		t.Fatalf("Load() error = %v, want ErrCorrupt", err)
	}
	if got := log.Entries(); len(got) != 1 || got[0].Text != "hello" {
		t.Fatalf("Load() entries = %#v, want the readable hello entry", got)
	}
}

func TestLoadOrQuarantineMovesCorruptFile(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	path := writeRawHistory(t, "libera", "#go", `{"server_time":1,"username":"alice","text":"hello"}`+"\n{broken")

	log, quarantined, err := LoadOrQuarantine("libera", "#go")
	if err != nil {
		t.Fatalf("LoadOrQuarantine() error = %v", err)
	}
	if !strings.HasPrefix(quarantined, path+".corrupt-") {
		t.Fatalf("quarantine path = %q, want prefix %q", quarantined, path+".corrupt-")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("corrupt file still at original path: %v", err)
	}
	if _, err := os.Stat(quarantined); err != nil {
		t.Fatalf("quarantined file missing: %v", err)
	}
	snapshot, dirty := log.Snapshot()
	if !dirty || len(snapshot.Entries) != 1 {
		t.Fatalf("recovered log dirty = %v with %d entries, want dirty with 1", dirty, len(snapshot.Entries))
	}
}

func TestLoadUnreadableFileIsNotCorrupt(t *testing.T) {
	skipIfRoot(t)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := FlushSnapshot("libera", "#go", []LogEntry{{ServerTime: 1, Text: "hello"}}); err != nil {
		t.Fatalf("FlushSnapshot() error = %v", err)
	}
	path := logFullPath("libera", "#go")
	if err := os.Chmod(path, 0); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, quarantined, err := LoadOrQuarantine("libera", "#go")
	if err == nil || errors.Is(err, ErrCorrupt) {
		t.Fatalf("LoadOrQuarantine() error = %v, want a non-corrupt read error", err)
	}
	if quarantined != "" {
		t.Fatalf("unreadable file was quarantined to %q", quarantined)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("unreadable file was moved: %v", err)
	}
}

func TestQuarantineOfMissingFileIsNotAnError(t *testing.T) {
	// A concurrent load or save may have already moved the corrupt file.
	quarantined, err := quarantine(filepath.Join(t.TempDir(), "missing.json.gz"))
	if err != nil || quarantined != "" {
		t.Fatalf("quarantine() = %q, %v; want empty path and no error", quarantined, err)
	}
}

func TestMergeSkipsDuplicateEntries(t *testing.T) {
	log := NewLog()
	log.Insert(LogEntry{ServerTime: 1, Username: "alice", Text: "old"})

	log.Merge([]LogEntry{
		{ServerTime: 1, Username: "alice", Text: "old"},
		{ServerTime: 2, Username: "bob", Text: "new"},
	})

	entries := log.Entries()
	if len(entries) != 2 || entries[0].Text != "old" || entries[1].Text != "new" {
		t.Fatalf("merged entries = %#v, want old then new", entries)
	}
}

// skipIfRoot skips permission-based tests, since root can read any file.
func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("file permissions are not enforced for root")
	}
}

func TestLogPathIgnoresCase(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	if logFullPath("libera", "#IdleRPG") != logFullPath("libera", "#idlerpg") {
		t.Fatal("different spellings of the same buffer should share a history file")
	}
}
