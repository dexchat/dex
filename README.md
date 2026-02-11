## dex

![preview](./preview.png)


## reference
* https://datatracker.ietf.org/doc/html/rfc2812
* https://modern.ircdocs.horse/

## TODO
* IRC:
  * enable scroll to viewport (chat or user list) based on the mouse is hovered
  * support for commands, e.g., /leave, /join, etc.
  * autocompletion for commands
* check if the chat being updated every message is a performance issue
* maybe centralize all paddings and borders numbers
* centralize and improve all styles in the styles package, make sure there are no loose styles
* implement command palette actions
* implement a new overlay: help with all the keybindings – will only be shown in the command palette
* migrate to lipgloss v2 and use their overlay
* fix mouse sequences being typed in input bar e.g. `[<64;76;47M[<65;76;47M`
* use all models as value and not reference
* create keybindings for scrolling the chat viewport
* new messages indicator
* tagged messages indicator
* improve channel list UI and user list member count, it's kinda ugly – ask for suggestions to someone
* improve ctrl+c quit color
* options that should be possible to set in the config file:
  * global nickname
  * display user join/part/quit messages
  * do not connect to the channels on startup
  * change the timestamp format and/or hide it
  * display channel modes
  * disable mouse/scroll support
* do not display the whole playback again on reconnection (I think it resets the buffer)
* display user modes in chat alongside the nickname
* display username in join/parts
* display nick in gray if user from history is not in channel
* match own nick color in input bar

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
