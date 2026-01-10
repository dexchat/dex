package irc

type UserListMsg struct {
	Server  string
	Channel string
	Users   []string
}

type BufferNewMessageMsg struct {
	Server  string
	Channel string
	Time    string
	From    string
	Text    string
}

type ChannelTopicMsg struct {
	Server  string
	Channel string
	Topic   string
}
