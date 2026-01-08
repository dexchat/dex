package config

type Config struct {
	Servers map[string]*Server `toml:"servers"`
}

type Server struct {
	Address  string   `toml:"address"`
	Port     int      `toml:"port"`
	SSL      bool     `toml:"ssl"`
	Password string   `toml:"password"`
	Channels []string `toml:"channels"`
	Nickname string   `toml:"nickname"`
	Username string   `toml:"username"`
	Realname string   `toml:"realname"`
}
