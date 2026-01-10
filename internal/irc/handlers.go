package irc

import (
	"log"

	"github.com/lrstanley/girc"
)

func (c *Client) onUserListChange(client *girc.Client, e girc.Event) {
	var channelName string
	if e.Command == girc.RPL_ENDOFNAMES {
		channelName = e.Params[1]
	} else {
		channelName = e.Params[0]
	}
	log.Printf("Handler triggered for event: %q, Source: %s, Params: %v", e.Command, e.Source, e.Params)
	channel := client.LookupChannel(channelName)
	if channel == nil {
		return
	}

	userList := make([]string, len(channel.UserList))

	for i, nick := range channel.UserList {
		prefix := ""
		user := client.LookupUser(nick)

		if perms, ok := user.Perms.Lookup(channelName); ok {
			switch {
			case perms.Owner:
				prefix = girc.OwnerPrefix
			case perms.Admin:
				prefix = girc.AdminPrefix
			case perms.Op:
				prefix = girc.OperatorPrefix
			case perms.HalfOp:
				prefix = girc.HalfOperatorPrefix
			case perms.Voice:
				prefix = girc.VoicePrefix
			}
		}

		userList[i] = prefix + nick
	}

	c.program.Send(UserListMsg{
		Server:  c.ServerName,
		Channel: channelName,
		Users:   userList,
	})
}
