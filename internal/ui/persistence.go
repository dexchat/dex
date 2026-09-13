package ui

import (
	"reflect"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/history"
)

type (
	historyFlushMsg         struct{}
	historyLoadedMsg        []loadedBufferHistory
	initialHistoryLoadedMsg struct {
		histories []loadedBufferHistory
		readState *history.ReadState
		readErr   error
	}
	loadedBufferHistory struct {
		key     BufferKey
		history *history.Log
	}
)

// historyFlushFinishedMsg reports the result of a persistence command back to
// the Bubble Tea update loop.
type historyFlushFinishedMsg struct {
	results           []historyFlushResult
	directMessagesErr error
	readStateErr      error
	directSnapshot    history.DirectMessages
	readStateSnapshot history.ReadState
	directDirty       bool
	readDirty         bool
}

type historyFlushResult struct {
	key      BufferKey
	revision uint64
	err      error
}

type historyFlushSnapshot struct {
	key      BufferKey
	server   string
	buffer   string
	snapshot history.LogSnapshot
}

type persistenceSnapshot struct {
	history        []historyFlushSnapshot
	directMessages history.DirectMessages
	readState      history.ReadState
	directDirty    bool
	readDirty      bool
}

func (s persistenceSnapshot) hasWork() bool {
	return len(s.history) > 0 || s.directDirty || s.readDirty
}

func (m *Model) snapshotPersistence() persistenceSnapshot {
	snapshot := persistenceSnapshot{
		directDirty: m.directMessagesDirty,
		readDirty:   m.readStateDirty,
	}
	if m.directMessages != nil {
		snapshot.directMessages.Users = append([]history.DirectMessage(nil), m.directMessages.Users...)
	}
	if m.readState != nil {
		snapshot.readState.Markers = make(map[string]history.ReadMarker, len(m.readState.Markers))
		for key, marker := range m.readState.Markers {
			snapshot.readState.Markers[key] = marker
		}
	}
	for key, buf := range m.buffers {
		logSnapshot, ok := buf.History.Snapshot()
		if !ok {
			continue
		}
		snapshot.history = append(snapshot.history, historyFlushSnapshot{
			key:      key,
			server:   buf.Server,
			buffer:   buf.Buffer,
			snapshot: logSnapshot,
		})
	}
	return snapshot
}

func (m *Model) startPersistence() tea.Cmd {
	if m.historyFlushInFlight {
		return nil
	}
	snapshot := m.snapshotPersistence()
	if !snapshot.hasWork() {
		return nil
	}
	m.historyFlushInFlight = true
	return flushHistoryCmd(snapshot)
}

func flushHistoryCmd(snapshot persistenceSnapshot) tea.Cmd {
	// The command only uses copied data, so it can perform disk I/O without
	// reading or mutating the live Bubble Tea model.
	return func() tea.Msg {
		results := make([]historyFlushResult, 0, len(snapshot.history))
		for _, item := range snapshot.history {
			err := history.FlushSnapshot(item.server, item.buffer, item.snapshot.Entries)
			results = append(results, historyFlushResult{
				key:      item.key,
				revision: item.snapshot.Revision,
				err:      err,
			})
		}
		result := historyFlushFinishedMsg{
			results:           results,
			directDirty:       snapshot.directDirty,
			readDirty:         snapshot.readDirty,
			directSnapshot:    snapshot.directMessages,
			readStateSnapshot: snapshot.readState,
		}
		if snapshot.directDirty {
			result.directMessagesErr = history.FlushDirectMessagesSnapshot(snapshot.directMessages)
		}
		if snapshot.readDirty {
			result.readStateErr = history.FlushReadStateSnapshot(snapshot.readState)
		}
		return result
	}
}

func (m *Model) applyPersistenceResult(msg historyFlushFinishedMsg) {
	m.historyFlushInFlight = false
	for _, result := range msg.results {
		if result.err != nil {
			continue
		}
		if buf := m.buffers[result.key]; buf != nil {
			buf.History.MarkPersisted(result.revision)
		}
	}
	if msg.directDirty && msg.directMessagesErr == nil && m.directMessages != nil && reflect.DeepEqual(*m.directMessages, msg.directSnapshot) {
		m.directMessagesDirty = false
	}
	if msg.readDirty && msg.readStateErr == nil && m.readState != nil && reflect.DeepEqual(*m.readState, msg.readStateSnapshot) {
		m.readStateDirty = false
	}
}
