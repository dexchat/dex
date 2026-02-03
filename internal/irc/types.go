package irc

import "time"

type UserListMsg struct {
	Server  string
	Channel string
	Users   []string
}

type BufferNewMessageMsg struct {
	Server    string
	Buffer    string
	Timestamp time.Time
	From      string
	Text      string
}

type ChannelTopicMsg struct {
	Server  string
	Channel string
	Topic   string
}

type NickUpdateMsg struct {
	Server string
	Nick   string
}

type ChannelNameUpdateMsg struct {
	Server        string
	CanonicalName string
}
