# UI Development Guide

This file supplements the repository-level `AGENTS.md`. Follow both files when
working under `internal/ui/`; this file provides the more specific guidance for
the TUI subtree.

## General guidelines

- Keep the Bubble Tea update loop responsive.
- Keep things simple; do not overcomplicate.
- Do not add network calls, disk scans, long loops, or expensive rendering
  directly to `Model.Update`.
- Represent asynchronous results as typed `tea.Msg` values and apply their
  state changes in `Update`.
- Use `tea.Cmd`, `tea.Batch`, or the established IRC goroutine boundary for
  side effects. Do not mutate UI model state from a command or goroutine.
- Keep message ownership clear: IRC code emits IRC messages, the top-level UI
  routes them, and components handle their own presentation state.
- Preserve input responsiveness during ZNC playback and large history loads.
- Use Lip Gloss width and height helpers for rendered terminal content; do not
  assume byte length equals terminal cell width.

## Architecture

### Top-level model

`ui.Model` in `ui.go` is the application-level Bubble Tea model. It owns:

- configuration and the IRC client manager;
- terminal size, focus state, and theme;
- the buffer map plus active and previous buffer keys;
- persisted read state and direct-message state;
- the channels sidebar, command palette, help overlay, and notification state.

Its `Update` method is the central router. It handles application and IRC
messages first, gives visible overlays priority, routes mouse-wheel events to
the pane under the pointer, and then updates the active chat, user list, and
channel list.

Do not bypass this routing casually. New cross-component behavior should
normally use a typed message handled by the top-level model.

### Buffers

`Buffer` in `buffer.go` represents a server, channel, or private-message view.
Each buffer owns its `chat.Model`, `users.Model`, `history.Log`, unread counters,
latest read marker, cached channel membership, and lazy-history flags.

- Build keys with `makeBufferKey`; IRC server and target names are
  case-insensitive.
- Use `getOrCreateBuffer` so creation stays synchronized with the channels
  sidebar and lazy-history setup.
- A server buffer has an empty `Buffer` field. `Buffer.isValid` is true only for
  sendable channel or private-message targets.
- Keep inactive channel membership as data. Render the user list when the
  buffer becomes active instead of eagerly rendering every playback update.

### Components

```text
components/
  channels/      Server and buffer tree, selection, badges, and notifications
  chat/          Topic, transcript viewport, input, formatting, and completion
  help/          Help overlay
  keybindings/   Shared key definitions and quit handling
  palette/       Searchable commands and channel selection
  users/         Channel member viewport and IRC prefix display
styles/          Themes, semantic style fields, and nickname color allocation
```

Child models use the return conventions already established in their package:
some are values and some are pointers. Always assign the model returned by
`Update` or `SetSize` where the API returns one.

### Rendering pipeline

1. `ui.Model.View` asks `layout.View` for the three-pane base view.
2. `layout.View` renders channels, the active chat, and users, then joins them
   horizontally.
3. A visible palette or help view is composited over the base with
   `overlayCenter`.
4. The resulting `tea.View` enables the alternate screen, focus reporting, and
   cell-motion mouse events.

Keep sizing centralized in the existing `calculate*` helpers and `SetSize`
methods. Always account for padding and borders in width calculations, as well
as topic and input heights. Handle very small terminals. Overlays must remain
clipped and centered without making the background exceed the terminal
dimensions.

## Message and side-effect rules

- Define small typed messages for events; avoid stringly typed routing.
- Commands may perform work and return messages, but model mutation belongs in
  `Update` or a synchronous helper called by it.
- Keep IRC socket handlers non-blocking. Large streams must remain batched or
  coalesced before reaching expensive UI work.
- Do not perform an entire playback burst in one update. Preserve
  `maxPlaybackMessagesPerUpdate` chunking unless measurements justify a change.
- Preserve the queued chat-rendering path: append with `QueueMessage`, schedule
  a flush, and render the active chat on `flushChatMsg`.
- History loading is intentionally lazy. Do not start one disk-loading job per
  discovered channel during connection playback.
- Do not insert an outgoing message into history before the IRC echo arrives;
  the echo supplies the authoritative server timestamp and message ID.
- Deduplication, latest-message tracking, unread counts, mentions, and
  notifications are separate responsibilities. Test each affected behavior
  when changing the incoming-message path.

## Input and overlays

- Palette and help overlays consume keyboard input while visible; do not allow
  keys to leak into the chat input or channel navigation.
- Mouse-wheel events are routed to channels, chat, or users based on terminal
  coordinates. Preserve this pane-specific behavior.
- Reuse `keybindings.DefaultKeyMap` instead of duplicating key definitions.
- External-editor work must use Bubble Tea process commands so terminal control
  is released and restored correctly. Temporary files must be cleaned up.

## Styling

- Read colors from `styles.Theme`; prefer semantic fields over colors hardcoded
  inside components.
- Keep complete theme palettes in `styles/colorschemes.go` and shared theme
  structure in `styles/theme.go`.
- When adding a semantic UI role, add the required theme field and populate it
  in every theme.
- Preserve terminal-cell-aware sizing when truncating, centering, or laying out
  styled strings, Unicode text, and IRC nick prefixes.
- Keep nickname coloring deterministic through `styles.UsernameColors`.

## Testing

Run the focused UI suite while iterating:

```sh
env GOCACHE=/tmp/dex-go-build go test ./internal/ui/...
```

For changes involving goroutines, IRC delivery, batching, playback, history, or
shared state, also run the repository-level race command from the root
`AGENTS.md`. Add focused tests beside the affected model or component,
including narrow terminal sizes, empty state, inactive buffers, overlay
visibility, and playback bursts when relevant.
