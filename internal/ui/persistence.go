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
	// shutdownRequested delays quitting until the final persistence command completes.
	shutdownRequested bool
	// pendingClose keeps a private buffer alive until its history is persisted.
	pendingClose BufferKey
	// channelRemovals prevents duplicate persistence commands for a parted channel.
	channelRemovals map[BufferKey]bool
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
	closeBufferFinishedMsg struct {
		key BufferKey
		err error
	}
	channelRemovalFinishedMsg struct {
		key       BufferKey
		server    string
		channel   string
		wasActive bool
		err       error
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
	if m.persistence.flushInFlight {
		return nil
	}
	snapshot := m.snapshotPersistence()
	if !snapshot.hasWork() {
		return nil
	}
	m.persistence.flushInFlight = true
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

func flushSingleHistoryCmd(key BufferKey, server, buffer string, snapshot history.LogSnapshot) tea.Cmd {
	return func() tea.Msg {
		return closeBufferFinishedMsg{
			key: key,
			err: history.FlushSnapshot(server, buffer, snapshot.Entries),
		}
	}
}

func flushChannelHistoryCmd(key BufferKey, server, channel string, wasActive bool, snapshot history.LogSnapshot) tea.Cmd {
	return func() tea.Msg {
		return channelRemovalFinishedMsg{
			key:       key,
			server:    server,
			channel:   channel,
			wasActive: wasActive,
			err:       history.FlushSnapshot(server, channel, snapshot.Entries),
		}
	}
}

func (m *Model) finishCloseBuffer(msg closeBufferFinishedMsg) {
	if m.persistence.pendingClose != msg.key {
		return
	}
	m.persistence.pendingClose = ""

	buffer, ok := m.buffers[msg.key]
	if !ok {
		return
	}
	if msg.err != nil {
		m.addCommandError(buffer, "error: could not close private message: "+msg.err.Error())
		return
	}

	delete(m.buffers, buffer.Key)
	if m.directMessages.Remove(buffer.Server, buffer.Buffer) {
		m.directMessagesDirty = true
	}
	m.activeBuffer = makeBufferKey(buffer.Server, "")
	m.channels, _ = m.channels.Update(channels.RemoveBufferMsg{
		Server:       buffer.Server,
		Buffer:       buffer.Buffer,
		SelectServer: true,
	})
}

func (m *Model) finishChannelRemoval(msg channelRemovalFinishedMsg) {
	delete(m.persistence.channelRemovals, msg.key)

	buffer, ok := m.buffers[msg.key]
	if !ok {
		return
	}
	if msg.err != nil {
		m.addCommandError(buffer, "error: could not save channel history: "+msg.err.Error())
		return
	}

	if msg.wasActive {
		m.activeBuffer = makeBufferKey(msg.server, "")
	}
	delete(m.buffers, msg.key)
	m.channels, _ = m.channels.Update(channels.RemoveBufferMsg{
		Server:       msg.server,
		Buffer:       msg.channel,
		SelectServer: msg.wasActive,
	})
}

func (m *Model) applyPersistenceResult(msg historyFlushFinishedMsg) {
	m.persistence.flushInFlight = false
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
