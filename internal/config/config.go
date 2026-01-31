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

// ConnectionUsername returns the username to use for connection
// Defaults to ConnectionNickname() if not set
func (s *Server) ConnectionUsername() string {
	if s.Username != "" {
		return s.Username
	}
	return s.ConnectionNickname()
}

// ConnectionPassword returns the server password to use for authentication
// For ZNC users, it constructs "nickname:password" format.
func (s *Server) ConnectionPassword() string {
	if s.isZNC() {
		return s.Nickname + ":" + s.Password
	}
	return s.Password
}
