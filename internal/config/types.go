package config

// rawConfig stores the config file as it is
type rawConfig struct {
	Servers map[string]*Server `toml:"servers"`
}

// Config stores the config file sorted and ready to be used
// Go have unordered maps which might cause inconsistencies
// in how the UI iterates over the servers and channels
type Config struct {
	Servers []*Server
}

type Server struct {
	Name     string
	Address  string   `toml:"address"`
	Port     int      `toml:"port"`
	SSL      bool     `toml:"ssl"`
	Password string   `toml:"password"`
	Channels []string `toml:"channels"`
	Nickname string   `toml:"nickname"`
	Username string   `toml:"username"`
	Realname string   `toml:"realname"`
}
