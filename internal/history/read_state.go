package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type ReadMarker struct {
	ServerTime int64  `json:"server_time"`
	MsgID      string `json:"msg_id,omitempty"`
}

type ReadState struct {
	Markers map[string]ReadMarker `json:"markers"`
}

func LoadReadState() (*ReadState, error) {
	state := &ReadState{
		Markers: make(map[string]ReadMarker),
	}

	file, err := os.Open(readStatePath())
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(state); err != nil {
		return nil, err
	}
	if state.Markers == nil {
		state.Markers = make(map[string]ReadMarker)
	}

	return state, nil
}

func (state *ReadState) Marker(server, buffer string) (ReadMarker, bool) {
	if state == nil {
		return ReadMarker{}, false
	}
	marker, ok := state.Markers[readMarkerKey(server, buffer)]
	return marker, ok
}

func (state *ReadState) MarkRead(server, buffer string, marker ReadMarker) {
	if state.Markers == nil {
		state.Markers = make(map[string]ReadMarker)
	}
	state.Markers[readMarkerKey(server, buffer)] = marker
}

func (state *ReadState) Flush() error {
	path := readStatePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, ".read-state-*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)

	if err := json.NewEncoder(file).Encode(state); err != nil {
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

func readStatePath() string {
	return filepath.Join(filepath.Dir(historyDir()), "read-state.json")
}

func readMarkerKey(server, buffer string) string {
	return strings.ToLower(server) + ":" + strings.ToLower(buffer)
}
