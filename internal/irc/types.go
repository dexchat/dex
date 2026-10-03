package irc

import "time"

// Event is a notification from an IRC client to the UI. Events from one
// server are delivered in the order its handlers queued them.
type Event interface {
	ircEvent()
}

func (UserListMsg) ircEvent()          {}
func (BufferNewMessageMsg) ircEvent()  {}
func (ChannelTopicMsg) ircEvent()      {}
func (ChannelTopicSetByMsg) ircEvent() {}
func (NickUpdateMsg) ircEvent()        {}
func (ChannelNameUpdateMsg) ircEvent() {}
func (ChannelJoinedMsg) ircEvent()     {}
func (ChannelPartedMsg) ircEvent()     {}

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
	Text          string // For an ACTION, the text without CTCP framing
	Action        bool   // True for a CTCP ACTION sent with /me
	MsgID         string // IRCv3 msgid, nil if the server does not support it
	OwnEcho       bool   // True if this is the own echo-message of a message sent from this client
	UserEvent     bool   // True for another user's JOIN, PART, QUIT, or KICK event
	Type          MessageType
}

// ChannelTopicMsg reports a channel's topic. A TOPIC change also reports who
// set it and when; RPL_TOPIC does not, and ChannelTopicSetByMsg follows it.
type ChannelTopicMsg struct {
	Server  string
	Channel string
	Topic   string
	SetBy   string
	SetAt   time.Time
}

// ChannelTopicSetByMsg reports who set a channel's current topic and when,
// from RPL_TOPICWHOTIME. SetAt is zero if the server sent no valid time.
type ChannelTopicSetByMsg struct {
	Server  string
	Channel string
	SetBy   string
	SetAt   time.Time
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

type ChannelPartedMsg struct {
	Server  string
	Channel string
}
