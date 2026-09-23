package history

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"
)

type Log struct {
	entries       []LogEntry
	lastUpdatedAt *time.Time
	revision      uint64
	persisted     uint64
}

// LogSnapshot is an immutable copy of a log at a specific revision.
type LogSnapshot struct {
	Entries  []LogEntry
	Revision uint64
}

func NewLog() *Log {
	return &Log{
		entries: make([]LogEntry, 0),
	}
}

func (log *Log) Insert(logEntry LogEntry) {
	// Find the insert position
	idx := sort.Search(len(log.entries), func(i int) bool {
		return log.entries[i].ServerTime > logEntry.ServerTime
	})

	log.entries = slices.Insert(log.entries, idx, logEntry)
	log.revision++

	now := time.Now()
	log.lastUpdatedAt = &now
}

// Merge inserts the entries that are not already in the log.
func (log *Log) Merge(entries []LogEntry) {
	for _, entry := range entries {
		if !log.IsDuplicate(entry) {
			log.Insert(entry)
		}
	}
}

// markDirty forces the log to be included in the next snapshot, e.g. after
// its entries were recovered from a file that no longer exists on disk.
func (log *Log) markDirty() {
	log.revision++
}

func (log *Log) Snapshot() (LogSnapshot, bool) {
	if log.revision == log.persisted {
		return LogSnapshot{}, false
	}
	entries := make([]LogEntry, len(log.entries))
	copy(entries, log.entries)
	return LogSnapshot{Entries: entries, Revision: log.revision}, true
}

// MarkPersisted records which revision was successfully written to disk.
// A newer revision remains dirty and will be included in a later snapshot.
func (log *Log) MarkPersisted(revision uint64) {
	if revision > log.persisted && revision <= log.revision {
		log.persisted = revision
		if revision == log.revision {
			log.lastUpdatedAt = nil
		}
	}
}

func (log *Log) Entries() []LogEntry {
	return log.entries
}

func (log *Log) IsDuplicate(entry LogEntry) bool {
	if len(log.entries) == 0 {
		return false
	}

	// Search within 1 second window
	windowNs := int64(time.Second)

	// Find start of window
	startTime := entry.ServerTime - windowNs
	startIdx := sort.Search(len(log.entries), func(i int) bool {
		return log.entries[i].ServerTime >= startTime
	})

	// Find end of window
	endTime := entry.ServerTime + windowNs
	endIdx := sort.Search(len(log.entries), func(i int) bool {
		return log.entries[i].ServerTime > endTime
	})

	// Check candidates in window
	for i := startIdx; i < endIdx; i++ {
		stored := log.entries[i]

		// If the server supports the msgid capability, we use it as the duplication
		// checking would be more accurate
		if entry.MsgID != nil && stored.MsgID != nil && *entry.MsgID == *stored.MsgID {
			return true
		}

		// Exact ServerTime + content match
		if stored.ServerTime == entry.ServerTime &&
			stored.Username == entry.Username &&
			stored.Text == entry.Text {
			return true
		}
	}

	return false
}

// ErrCorrupt reports a history file that exists but cannot be decoded.
var ErrCorrupt = errors.New("corrupt history file")

// Load reads the stored history for a buffer. A missing file is an empty
// log. A corrupt file returns the entries decoded before the damage along
// with an error wrapping ErrCorrupt. Any other error means the file could not
// be read and must not be replaced.
func Load(server, buffer string) (*Log, error) {
	l := NewLog()

	path := logFullPath(server, buffer)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return l, nil
	}
	if err != nil {
		return l, fmt.Errorf("open history: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return l, fmt.Errorf("%w: %s: %w", ErrCorrupt, path, err)
	}
	defer gz.Close()

	decoder := json.NewDecoder(gz)
	for {
		var entry LogEntry
		err := decoder.Decode(&entry)
		if errors.Is(err, io.EOF) {
			return l, nil
		}
		if err != nil {
			return l, fmt.Errorf("%w: %s: %w", ErrCorrupt, path, err)
		}
		l.entries = append(l.entries, entry)
	}
}

// LoadOrQuarantine loads a buffer's history. When the file is corrupt, it
// renames the file to "<path>.corrupt-<timestamp>" and returns the readable
// entries marked dirty, so the next save writes them back without destroying
// the original. It returns the quarantine path when this call moved the file.
func LoadOrQuarantine(server, buffer string) (*Log, string, error) {
	log, err := Load(server, buffer)
	if !errors.Is(err, ErrCorrupt) {
		return log, "", err
	}
	quarantined, qerr := quarantine(logFullPath(server, buffer))
	if qerr != nil {
		return log, "", fmt.Errorf("%w; quarantine: %w", err, qerr)
	}
	log.markDirty()
	return log, quarantined, nil
}

// quarantine moves a corrupt history file aside. A file that is already gone
// was moved by a concurrent load or save, which reports it instead, so this
// returns an empty path without error.
func quarantine(path string) (string, error) {
	target := fmt.Sprintf("%s.corrupt-%s", path, time.Now().Format("20060102T150405.000000000"))
	err := os.Rename(path, target)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return target, nil
}

func (log *Log) Flush(server, buffer string) error {
	if log.lastUpdatedAt == nil {
		return nil
	}
	if err := FlushSnapshot(server, buffer, log.entries); err != nil {
		return err
	}
	log.persisted = log.revision
	log.lastUpdatedAt = nil
	return nil
}

func FlushSnapshot(server, buffer string, entries []LogEntry) error {
	path := logFullPath(server, buffer)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, ".history-*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(tempPath)
	}

	gz := gzip.NewWriter(file)
	encoder := json.NewEncoder(gz)
	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			_ = gz.Close()
			cleanup()
			return err
		}
	}
	if err := gz.Close(); err != nil {
		cleanup()
		return err
	}
	if err := file.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}
