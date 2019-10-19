// Package config holds the configuration struct.
package config

type Config struct {
	ServeAddress   string
	MusicDirectory string
	CacheDirectory string
}

// Default returns the default config.
func Default() *Config {
	conf := &Config{
		ServeAddress: "127.0.0.1:8118",
	}
	return conf
}
