package config

import "strings"

// UseSSL returns whether SSL should be used. Defaults to true if not set.
func (s *Server) UseSSL() bool {
	if s.SSL == nil {
		return true
	}
	return *s.SSL
}

// isBouncer reports whether the nickname uses the bouncer login format
// "user/network". A client name may be added as "user/network@client" for
// soju or "user@client/network" for ZNC. Nicknames cannot contain "/", so
// this cannot match a real nickname.
func (s *Server) isBouncer() bool {
	return strings.Contains(s.Nickname, "/")
}

// ConnectionNickname returns the nickname to use for display and connection
// For bouncer logins, it is the user part before the network or client.
func (s *Server) ConnectionNickname() string {
	if s.isBouncer() {
		return s.Nickname[:strings.IndexAny(s.Nickname, "/@")]
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
// For bouncer logins, it is "login:password", the format ZNC reads from PASS.
func (s *Server) ConnectionPassword() string {
	if s.isBouncer() {
		return s.Nickname + ":" + s.Password
	}
	return s.Password
}

// SASLCredentials returns the SASL PLAIN username and password for bouncer
// logins, which soju requires: it reads the password from PASS as is. When
// SASL succeeds, soju ignores PASS; ZNC without SASL support uses PASS.
func (s *Server) SASLCredentials() (username, password string, ok bool) {
	if !s.isBouncer() {
		return "", "", false
	}
	return s.Nickname, s.Password, true
}
