package config

import "strings"

// UseSSL returns whether SSL should be used. Defaults to true if not set.
func (s *Server) UseSSL() bool {
	if s.SSL == nil {
		return true
	}
	return *s.SSL
}

// isZNC returns true if the nickname contains a "/" indicating ZNC format (user/network)
func (s *Server) isZNC() bool {
	return strings.Contains(s.Nickname, "/")
}

// ConnectionNickname returns the nickname to use for display and connection
// For ZNC users, it derives the nick from the part before "/"
func (s *Server) ConnectionNickname() string {
	if s.isZNC() {
		return strings.Split(s.Nickname, "/")[0]
	}
	return s.Nickname
}

// usesSASL reports whether the username carries a bouncer network or client
// suffix, as in soju's "user/network@client". Such names are not valid
// idents, so they are sent with SASL PLAIN and USER gets the bare username.
func (s *Server) usesSASL() bool {
	return strings.ContainsAny(s.Username, "/@")
}

// ConnectionUsername returns the username to use for connection
// Defaults to ConnectionNickname() if not set
// For bouncer usernames, it is the part before the network or client suffix.
func (s *Server) ConnectionUsername() string {
	if s.usesSASL() {
		return s.Username[:strings.IndexAny(s.Username, "/@")]
	}
	if s.Username != "" {
		return s.Username
	}
	return s.ConnectionNickname()
}

// ConnectionPassword returns the server password to use for authentication
// For ZNC users, it constructs "nickname:password" format.
// It is empty when the password is sent with SASL.
func (s *Server) ConnectionPassword() string {
	if s.usesSASL() {
		return ""
	}
	if s.isZNC() {
		return s.Nickname + ":" + s.Password
	}
	return s.Password
}

// SASLCredentials returns the SASL PLAIN username and password. ok is false
// when the server authenticates with PASS.
func (s *Server) SASLCredentials() (username, password string, ok bool) {
	if !s.usesSASL() {
		return "", "", false
	}
	return s.Username, s.Password, true
}
