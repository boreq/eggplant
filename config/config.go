// Package config holds the configuration struct.
package config

type Config struct {
	Debug        bool
	ServeAddress string
}

// Default returns the default config.
func Default() *Config {
	conf := &Config{
		Debug:        false,
		ServeAddress: "127.0.0.1:8118",
	}
	return conf
}
