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
	return filepath.Join(osDataDir, "dexchat", "history")
}

func logFullPath(server, buffer string) string {
	serverName := sanitize(server)
	channelHash := hashString(buffer)
	return filepath.Join(historyDir(), serverName, channelHash+".json.gz")
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return r.Replace(s)
}

func hashString(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%d", h.Sum64())
}
