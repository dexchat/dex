## dex

![preview](./preview.png)


## reference
* https://datatracker.ietf.org/doc/html/rfc2812
* https://modern.ircdocs.horse/

## config

```toml
[ui]
unread_badges = true
mention_badges = true

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go", "#random"]

# optional per-server overrides
unread_badges = false
mention_badges = true
```

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
* FIX: lag when loading channels in the channel list
* Think what should be included in the 0.1.0 so I can open the code


## 0.1.0
* config file with minimal options
* support for /leave, /join, /quit, at least


## Known bugs
* Users with non-standard channel prefixes (e.g., `!` for channel admin/creator) may not appear in the user list. This is a [girc](https://github.com/lrstanley/girc) limitation — its NAMES reply parser only recognizes `~ & @ % +` as prefixes, ignoring any additional ones the server advertises via `ISUPPORT`. Users with unrecognized prefixes are silently dropped from the channel state.

## Features to implement
* command palette with: quick jump, fuzzy channel search, join, leave, etc.
* ability to open vim in input to edit the message being sent
* spell checking in input
* option to show only channels with unread messages
* a new pane on top of the channel list with: buffer for all your DMs and tagged messages
* channel topic: config option to truncate to one line
* config file where user can set: themes, servers, etc.
* support natively most used weechat plugins: exec, alias, spell, logger, etc
* multiple lines paste confirmation -> option to upload it to pastebin and then send the link
* sound notification support
* vim mode
