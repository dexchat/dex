package ui

import (
	"reflect"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/history"
	"github.com/dexchat/dex/internal/ui/components/channels"
)

type persistenceState struct {
	// flushInFlight prevents overlapping persistence commands.
	flushInFlight bool
	// flushPending requests another snapshot after the current command finishes.
	flushPending bool
	// shutdownRequested delays quitting until the final persistence command
	// completes.
	shutdownRequested bool
	// detachedHistory keeps snapshots for buffers already removed from the UI.
	detachedHistory map[BufferKey]detachedHistory
}

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
	log      *history.Log
	revision uint64
	err      error
}

type historyFlushSnapshot struct {
	key      BufferKey
	server   string
	buffer   string
	log      *history.Log
	snapshot history.LogSnapshot
}

type detachedHistory struct {
	server string
	buffer string
	log    *history.Log
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
	for key, detached := range m.persistence.detachedHistory {
		logSnapshot, ok := detached.log.Snapshot()
		if !ok {
			continue
		}
		snapshot.history = append(snapshot.history, historyFlushSnapshot{
			key:      key,
			server:   detached.server,
			buffer:   detached.buffer,
			log:      detached.log,
			snapshot: logSnapshot,
		})
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
			log:      buf.History,
			snapshot: logSnapshot,
		})
	}
	return snapshot
}

func (m *Model) startPersistence() tea.Cmd {
	if m.persistence.flushInFlight {
		m.persistence.flushPending = true
		return nil
	}
	snapshot := m.snapshotPersistence()
	if !snapshot.hasWork() {
		return nil
	}
	m.persistence.flushInFlight = true
	m.persistence.flushPending = false
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
				log:      item.log,
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

func (m *Model) removeBuffer(buffer *Buffer, forgetDirectMessage bool) tea.Cmd {
	key := buffer.Key
	wasActive := m.activeBuffer == key
	if wasActive {
		m.markBufferRead(buffer)
		m.activeBuffer = makeBufferKey(buffer.Server, "")
	}
	if _, dirty := buffer.History.Snapshot(); dirty {
		m.persistence.detachedHistory[key] = detachedHistory{
			server: buffer.Server,
			buffer: buffer.Buffer,
			log:    buffer.History,
		}
	}
	delete(m.buffers, key)
	if forgetDirectMessage && m.directMessages.Remove(buffer.Server, buffer.Buffer) {
		m.directMessagesDirty = true
	}
	m.channels, _ = m.channels.Update(channels.RemoveBufferMsg{
		Server:       buffer.Server,
		Buffer:       buffer.Buffer,
		SelectServer: wasActive,
	})
	return m.startPersistence()
}

func (m *Model) applyPersistenceResult(msg historyFlushFinishedMsg) tea.Cmd {
	m.persistence.flushInFlight = false
	for _, result := range msg.results {
		if result.err != nil {
			continue
		}
		result.log.MarkPersisted(result.revision)
		if pending, ok := m.persistence.detachedHistory[result.key]; ok && pending.log == result.log {
			if _, dirty := pending.log.Snapshot(); !dirty {
				delete(m.persistence.detachedHistory, result.key)
			}
		}
	}
	if msg.directDirty && msg.directMessagesErr == nil && m.directMessages != nil && reflect.DeepEqual(*m.directMessages, msg.directSnapshot) {
		m.directMessagesDirty = false
	}
	if msg.readDirty && msg.readStateErr == nil && m.readState != nil && reflect.DeepEqual(*m.readState, msg.readStateSnapshot) {
		m.readStateDirty = false
	}

	retry := m.persistence.flushPending
	failed := msg.directMessagesErr != nil || msg.readStateErr != nil
	for _, result := range msg.results {
		failed = failed || result.err != nil
	}
	if m.persistence.shutdownRequested {
		if failed {
			return tea.Quit
		}
		retry = retry || m.snapshotPersistence().hasWork()
		if !retry {
			return tea.Quit
		}
	}
	if retry {
		return m.startPersistence()
	}
	return nil
}
