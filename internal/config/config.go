// Package config holds the configuration struct.
package config

import (
	"io"

	"github.com/pelletier/go-toml"
)

type ExposedConfig struct {
	ServeAddress string `toml:"serve_address" comment:"Specifies under which address you will be able to access the UI. The\n addresses are specified as \"ip:port\". If you want to listen only to\n local connections use \"127.0.0.1:XXXX\" as the IP and replace XXXX\n with a desired port. If you want to listen externally use\n \"0.0.0.0:XXXX\" as the IP and replace XXXX with a desired port."`

	MusicDirectory string `toml:"music_directory" comment:"Path to a directory containing your music."`

	DataDirectory string `toml:"data_directory" comment:"Path to a directory which will be used for data storage. Eggplant will store\n its database in this directory. This directory should never be purged."`

	CacheDirectory string `toml:"cache_directory" comment:"Path to a directory which will be used for caching converted tracks and\n thumbnails. You should not remove files from this directory unless necessary\n as Eggplant ensures that old data is automatically removed and removing the\n cached files will force Eggplant to convert all tracks and thumbnails again."`
}

type Config struct {
	ExposedConfig

	// Files with those extensions are recognized as tracks. Extensions in
	// this list should begin with a dot. Extensions are case insenitive.
	TrackExtensions []string

	// Files with those stems in their names and one of the thumbail
	// extensions are recognized as thumbnails. For example if "thumbnail"
	// is present in this list then files such as "thumbnail.jpg",
	// "thumbnail.png" etc. would be considered to be a thumbnail given
	// that ".jpg" and ".png" are present in the list of thumbnail
	// extensions. Stems are case insensitive.
	ThumbnailStems []string

	// Files with those extensions and one of the thumbnail stems in their
	// name are recognized as thumbnails. Extensions in this list should
	// begin with a dot. Extensions are case insenitive.
	ThumbnailExtensions []string
}

// Default returns the default config.
func Default() *Config {
	conf := &Config{
		ExposedConfig: ExposedConfig{
			ServeAddress:   "127.0.0.1:8118",
			MusicDirectory: "/path/to/music",
			DataDirectory:  "/path/to/data",
			CacheDirectory: "/path/to/cache",
		},
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
		ThumbnailExtensions: []string{
			".jpg",
			".jpeg",
			".png",
			".gif",
		},
	}
	return conf
}

func Marshal(w io.Writer, cfg ExposedConfig) error {
	return toml.NewEncoder(w).Encode(cfg)
}

func Unmarshal(r io.Reader, cfg *ExposedConfig) error {
	return toml.NewDecoder(r).Decode(cfg)
}
