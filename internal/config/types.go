package config

// rawConfig stores the config file as it is
type rawConfig struct {
	Servers map[string]*Server `toml:"servers"`
}

// Config stores the config file sorted and ready to be used
type Config struct {
	// Servers is used to create the buffers on startup only
	// Later, the source of truth is the irc client channels
	// as the user can join channels while using the app
	Servers []*Server
}

type Server struct {
	Name     string
	Address  string   `toml:"address"`
	Port     int      `toml:"port"`
	SSL      *bool    `toml:"ssl"`
	Password string   `toml:"password"`
	Channels []string `toml:"channels"`
	Nickname string   `toml:"nickname"`
	Username string   `toml:"username"`
	Realname string   `toml:"realname"`
}
