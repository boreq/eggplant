package commands

import (
	"context"
	"os"

	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/boreq/errors"
	"github.com/boreq/guinea"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "config",
			Optional:    false,
			Multiple:    false,
			Description: "Path to a configuration file",
		},
	},
	ShortDescription: "serves your music",
}

func runRun(c guinea.Context) error {
	conf, err := loadConfig(c.Arguments[0])
	if err != nil {
		return errors.Wrap(err, "could not load the configuration")
	}

	if err := validateConfig(conf); err != nil {
		return errors.Wrap(err, "problem with the configuration")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service, err := wire.BuildService(ctx, conf)
	if err != nil {
		return errors.Wrap(err, "could not create a service")
	}

	return service.Run(ctx)
}

func loadConfig(path string) (*config.Config, error) {
	conf := config.Default()

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the config file")
	}

	if err := config.Unmarshal(f, &conf.ExposedConfig); err != nil {
		return nil, errors.Wrap(err, "failed to decode the config")
	}

	return conf, nil
}

// validateConfig makes sure that the specific directories exist early on to
// avoid confusing the user with a ton of error messages being printed by the
// program at the later stages of execution
func validateConfig(conf *config.Config) error {
	if err := validateDirectory(conf.CacheDirectory); err != nil {
		return errors.Wrap(err, "problem with the cache directory")
	}

	if err := validateDirectory(conf.DataDirectory); err != nil {
		return errors.Wrap(err, "problem with the data directory")
	}

	if err := validateDirectory(conf.MusicDirectory); err != nil {
		return errors.Wrap(err, "problem with the music directory")
	}

	return nil
}

func validateDirectory(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return errors.Wrap(err, "stat returned an error")
	}
	return nil
}
