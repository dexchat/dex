package irc

type UserListMsg struct {
	Server  string
	Channel string
	Users   []string
}
