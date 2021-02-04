// Package config holds the configuration struct.
package config

type Config struct {
	ServeAddress   string
	MusicDirectory string
	DataDirectory  string

	// Files with those extensions are recognized as tracks. Extensions in
	// this list should begin with a dot.
	TrackExtensions []string
}

// Default returns the default config.
func Default() *Config {
	conf := &Config{
		ServeAddress: "127.0.0.1:8118",
		TrackExtensions: []string{
			".flac",
			".mp3",
			".ogg",
			".aac",
			".wav",
			".wma",
			".aiff",
		},
	}
	return conf
}
