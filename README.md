<h1 align="center">
  <img src="./images/dexchat-icon.png" alt="Dex logo" width="128">
  <br>
  dexchat
</h1>

<p align="center">
  A modern and fast TUI IRC client for the terminal, with quick channel
  jumping, command palettes, and more.
</p>

![preview](./images/preview.png)

## reference

* <https://datatracker.ietf.org/doc/html/rfc2812>
* <https://modern.ircdocs.horse/>

## TODO

* IRC:
  * support for commands, e.g., /leave, /join, etc.
  * autocompletion for commands
* check if the chat being updated every message is a performance issue
* maybe centralize all paddings and borders numbers
* implement command palette actions
* implement a new overlay: help with all the keybindings – will only be shown in the command palette
* use all models as value and not reference
* new messages indicator
* tagged messages indicator
* improve channel list UI and user list member count, it's kinda ugly – ask for suggestions to someone
* options that should be possible to set in the config file:
  * global nickname
  * display user join/part/quit messages
  * do not connect to the channels on startup
  * change the timestamp format and/or hide it
  * display channel modes
  * disable mouse/scroll support
* Think what should be included in the 0.1.0 so I can open the code

## 0.1.0

TODO:

* Usar features novas do charm v2? Composite? Trees?
DONE:
* notifications when mentioned
* nickname complete on tab
* palette and channel picker
* support for /leave, /join, /list, /msg, /part, /close
* notifications for the active buffer when the terminal is unfocused

## Features to implement

* command palette with: quick jump, fuzzy channel search, join, leave, etc.
* ability to open $EDITOR in input to edit the message being sent
* spell checking in input
* option to show only channels with unread messages
* a new pane on top of the channel list with: buffer for all your DMs and tagged messages
* channel topic: config option to truncate to one line
* config file where user can set: themes, servers, etc.
* support natively most used weechat plugins: exec, alias, spell, logger, etc
* multiple lines paste confirmation -> option to upload it to pastebin and then send the link
* sound notification support
* vim mode
* choose notification type
* notify using <https://github.com/DaltonSW/BubbleUp>

### Terminal focus notifications

Dex uses terminal focus events to decide whether notifications for the active
buffer should be suppressed. When running Dex inside tmux, enable event
forwarding in `~/.tmux.conf`:

```tmux
set -g focus-events on
```

Apply it to the current tmux server with `tmux set-option -g focus-events on`,
or restart tmux after updating the configuration.
