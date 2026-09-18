package history

import (
	"compress/gzip"
	"encoding/json"
	"errors"
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

func Load(server, buffer string) *Log {
	l := NewLog()

	path := logFullPath(server, buffer)
	file, err := os.Open(path)
	if err != nil {
		return l
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return l
	}
	defer gz.Close()

	decoder := json.NewDecoder(gz)
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		l.entries = append(l.entries, entry)
	}

	return l
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

// MergeStoredEntries combines incoming entries with an existing history file.
// It is used when live messages arrive before a buffer's history is loaded.
func MergeStoredEntries(server, buffer string, incoming []LogEntry) ([]LogEntry, error) {
	path := logFullPath(server, buffer)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return append([]LogEntry(nil), incoming...), nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	merged := NewLog()
	decoder := json.NewDecoder(gz)
	for {
		var entry LogEntry
		err := decoder.Decode(&entry)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		merged.Insert(entry)
	}
	for _, entry := range incoming {
		if !merged.IsDuplicate(entry) {
			merged.Insert(entry)
		}
	}
	return append([]LogEntry(nil), merged.Entries()...), nil
}
