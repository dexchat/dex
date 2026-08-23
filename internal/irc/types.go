package irc

import "time"

type UserListMsg struct {
	Server   string
	Channel  string
	Users    []string
	Prefixes string
}

type MessageType int

const (
	MessageTypeNormal MessageType = iota
	MessageTypeConnected
	MessageTypeDisconnected
	MessageTypeServer
)

type BufferNewMessageMsg struct {
	Server        string
	Buffer        string
	DirectMessage bool // True if IRC target is a user
	Timestamp     time.Time
	From          string
	Text          string
	MsgID         string // IRCv3 msgid, nil if the server does not support it
	OwnEcho       bool   // True if this is the own echo-message of a message sent from this client
	UserEvent     bool   // True for another user's JOIN, PART, or QUIT event
	Type          MessageType
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

type ChannelJoinedMsg struct {
	Server  string
	Channel string
}

type ChannelJoinedBatchMsg []ChannelJoinedMsg

type ChannelPartedMsg struct {
	Server  string
	Channel string
}

type BufferNewMessageBatchMsg []BufferNewMessageMsg
