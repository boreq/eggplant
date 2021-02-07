// Package config holds the configuration struct.
package config

type Config struct {
	// Specifies under which address you will be able to access the UI. The
	// addresses are specified as "ip:port". If you want to listen only to
	// local connections use "127.0.0.1:XXXX" as the IP and replace XXXX
	// with a desired port. If you want to listen externally use
	// "0.0.0.0:XXXX" as the IP and replace XXX with a desired port.
	ServeAddress string

	// Path to a directory containing your music.
	MusicDirectory string

	// Path to a directory which will be used for data storage. Eggplant
	// will store its database and converted files in this directory.
	DataDirectory string

	// Files with those extensions are recognized as tracks. Extensions in
	// this list should begin with a dot. Extensions are case insenitive.
	TrackExtensions []string

	// Files with those stems in their names are recognized as thumbnails.
	// For example if "thumbnail" is present in this list then files such
	// as "thumbnail.jpg", "thumbnail.png" etc. would be considered a
	// thumbnail. Stems are case insensitive.
	ThumbnailStems []string
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
			".opus",
		},
		ThumbnailStems: []string{
			"thumbnail",
			"album",
			"cover",
			"folder",
		},
	}
	return conf
}
