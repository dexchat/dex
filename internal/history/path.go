package history

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
)

func historyDir() string {
	osDataDir := os.Getenv("XDG_DATA_HOME")
	if osDataDir == "" {
		home, _ := os.UserHomeDir()
		osDataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(osDataDir, "dex", "history")
}

// logFullPath returns the history file for a buffer. IRC names are
// case-insensitive, so the name goes through NameKey like buffer keys and
// read markers.
func logFullPath(server, buffer string) string {
	return filepath.Join(historyDir(), sanitize(server), hashString(NameKey(buffer))+".json.gz")
}

func sanitize(s string) string {
	s = NameKey(s)
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return r.Replace(s)
}

func hashString(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%d", h.Sum64())
}
