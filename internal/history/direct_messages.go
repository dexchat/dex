package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type DirectMessage struct {
	Server string `json:"server"`
	User   string `json:"user"`
}

type DirectMessages struct {
	Users []DirectMessage `json:"direct_messages"`
}

func LoadDirectMessages() (*DirectMessages, error) {
	directMessages := &DirectMessages{}
	file, err := os.Open(directMessagesPath())
	if errors.Is(err, os.ErrNotExist) {
		return directMessages, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(directMessages); err != nil {
		return nil, err
	}
	return directMessages, nil
}

func (d *DirectMessages) Add(server, user string) bool {
	for _, directMessage := range d.Users {
		if strings.EqualFold(directMessage.Server, server) && strings.EqualFold(directMessage.User, user) {
			return false
		}
	}
	d.Users = append(d.Users, DirectMessage{Server: server, User: user})
	return true
}

func (d *DirectMessages) Flush() error {
	path := directMessagesPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".direct-messages-*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if err := json.NewEncoder(file).Encode(d); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func directMessagesPath() string {
	return filepath.Join(filepath.Dir(historyDir()), "direct-messages.json")
}
