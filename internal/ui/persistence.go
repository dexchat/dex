package ui

import (
	"errors"
	"fmt"
	"reflect"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/history"
	"github.com/dexchat/dex/internal/ui/components/channels"
)

// persistenceState coordinates asynchronous saves and keeps removed buffers
// available until their pending history has been written.
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

// historyFlushFinishedMsg carries persistence results back to the Bubble Tea
// update loop, where dirty flags and shutdown state are updated safely.
type (
	historyFlushMsg         struct{}
	historyLoadedMsg        []loadedBufferHistory
	initialHistoryLoadedMsg struct {
		histories []loadedBufferHistory
		readState *history.ReadState
		readErr   error
	}
	loadedBufferHistory struct {
		key         BufferKey
		history     *history.Log
		quarantined string
		err         error
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

// historyFlushResult reports whether one history snapshot was written and
// which log revision was included in that write.
type historyFlushResult struct {
	key      BufferKey
	log      *history.Log
	revision uint64
	err      error
	// readErr means the stored history could not be read for merging, so
	// nothing was written and the buffer must stop persisting.
	readErr error
	// quarantined is the path a corrupt stored history was moved to.
	quarantined string
}

// historyFlushSnapshot contains immutable data for one asynchronous history
// write.
type historyFlushSnapshot struct {
	key      BufferKey
	server   string
	buffer   string
	log      *history.Log
	snapshot history.LogSnapshot

	// mergeExisting is true when the live log does not include disk history.
	mergeExisting bool
}

// detachedHistory retains a removed buffer's log until its pending write
// completes, including whether disk history must be merged first.
type detachedHistory struct {
	server        string
	buffer        string
	log           *history.Log
	mergeExisting bool
}

// persistenceSnapshot is the complete set of copied data used by one save
// command. It must not reference mutable collections while the command runs.
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
			key:           key,
			server:        detached.server,
			buffer:        detached.buffer,
			log:           detached.log,
			snapshot:      logSnapshot,
			mergeExisting: detached.mergeExisting,
		})
	}
	for key, buf := range m.buffers {
		if buf.historyState == historyUnreadable {
			continue
		}
		logSnapshot, ok := buf.History.Snapshot()
		if !ok {
			continue
		}
		snapshot.history = append(snapshot.history, historyFlushSnapshot{
			key:           key,
			server:        buf.Server,
			buffer:        buf.Buffer,
			log:           buf.History,
			snapshot:      logSnapshot,
			mergeExisting: buf.historyState != historyLoaded,
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

// requestShutdown saves pending local data and quits once it is written. A
// failed save keeps the app open for a retry; see applyPersistenceResult.
func (m *Model) requestShutdown() tea.Cmd {
	m.persistence.shutdownRequested = true
	m.markBufferRead(m.getActiveBuffer())
	if cmd := m.startPersistence(); cmd != nil {
		return cmd
	}
	if m.persistence.flushInFlight {
		return nil
	}
	return tea.Quit
}

func flushHistoryCmd(snapshot persistenceSnapshot) tea.Cmd {
	// The command only uses copied data, so it can perform disk I/O without
	// reading or mutating the live Bubble Tea model.
	return func() tea.Msg {
		results := make([]historyFlushResult, 0, len(snapshot.history))
		for _, item := range snapshot.history {
			result := historyFlushResult{
				key:      item.key,
				log:      item.log,
				revision: item.snapshot.Revision,
			}
			entries := item.snapshot.Entries
			if item.mergeExisting {
				stored, quarantined, err := history.LoadOrQuarantine(item.server, item.buffer)
				if err != nil {
					result.readErr = err
					results = append(results, result)
					continue
				}
				result.quarantined = quarantined
				stored.Merge(entries)
				entries = stored.Entries()
			}
			result.err = history.FlushSnapshot(item.server, item.buffer, entries)
			results = append(results, result)
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
		// The new active buffer may not have been resized since it last lost
		// focus (WindowSizeMsg only resizes the currently active buffer), so
		// bring it up to date the way selectBuffer does on a normal switch.
		if newActive, ok := m.buffers[m.activeBuffer]; ok {
			newActive.Chat.SetSize(m.calculateChatWidth(), m.calculateChatHeight())
			newActive.Users = newActive.Users.SetSize(usersPanelMaxWidth, m.calculateChatHeight())
		}
	}
	if _, dirty := buffer.History.Snapshot(); dirty && buffer.historyState != historyUnreadable {
		m.persistence.detachedHistory[key] = detachedHistory{
			server:        buffer.Server,
			buffer:        buffer.Buffer,
			log:           buffer.History,
			mergeExisting: buffer.historyState != historyLoaded,
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
		if result.quarantined != "" {
			m.addCommandError(m.bufferForFlushResult(result), quarantineNotice(result.quarantined))
		}
		if result.readErr != nil {
			m.disableHistoryPersistence(result)
			continue
		}
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
	persistenceErr := persistenceResultError(msg)
	if m.persistence.shutdownRequested {
		if persistenceErr != nil {
			m.persistence.shutdownRequested = false
			m.persistence.flushPending = false
			m.addCommandError(m.getActiveBuffer(), "error: could not save local data; press ctrl+c to retry: "+persistenceErr.Error())
			return nil
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

// disableHistoryPersistence stops saving the log whose stored history could
// not be read during a save, so a later save cannot replace that file.
func (m *Model) disableHistoryPersistence(result historyFlushResult) {
	if buf, ok := m.buffers[result.key]; ok && buf.History == result.log {
		m.markHistoryUnreadable(buf, result.readErr)
		return
	}
	if pending, ok := m.persistence.detachedHistory[result.key]; ok && pending.log == result.log {
		delete(m.persistence.detachedHistory, result.key)
		m.addCommandError(m.getActiveBuffer(), historyReadErrorNotice(result.readErr))
	}
	// Otherwise the log was replaced by a completed load, which merged its
	// entries, so the failed save is stale.
}

// markHistoryUnreadable disables saves for buf and reports the error once.
func (m *Model) markHistoryUnreadable(buf *Buffer, err error) {
	if buf.historyState != historyUnreadable {
		m.addCommandError(buf, historyReadErrorNotice(err))
	}
	buf.historyState = historyUnreadable
}

// bufferForFlushResult returns the buffer that owns a saved log, or the
// active buffer when that buffer was already removed.
func (m *Model) bufferForFlushResult(result historyFlushResult) *Buffer {
	if buf, ok := m.buffers[result.key]; ok && buf.History == result.log {
		return buf
	}
	return m.getActiveBuffer()
}

func historyReadErrorNotice(err error) string {
	return "error: could not read history; it will not be saved until it can be read: " + err.Error()
}

func quarantineNotice(path string) string {
	return "warning: history file was corrupt and moved to " + path
}

func persistenceResultError(msg historyFlushFinishedMsg) error {
	var errs []error
	for _, result := range msg.results {
		if result.err != nil {
			errs = append(errs, fmt.Errorf("history %s: %w", result.key, result.err))
		}
	}
	if msg.directMessagesErr != nil {
		errs = append(errs, fmt.Errorf("direct messages: %w", msg.directMessagesErr))
	}
	if msg.readStateErr != nil {
		errs = append(errs, fmt.Errorf("read state: %w", msg.readStateErr))
	}
	return errors.Join(errs...)
}
