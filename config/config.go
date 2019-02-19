// Package config holds the configuration struct.
package config

type ConfigStruct struct {
	Debug        bool
	ServeAddress string
}

// Default returns the default config.
func Default() *ConfigStruct {
	conf := &ConfigStruct{
		Debug:        false,
		ServeAddress: "127.0.0.1:8118",
	}
	return conf
}
