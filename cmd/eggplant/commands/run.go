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

	service, err := wire.BuildService(conf)
	if err != nil {
		return errors.Wrap(err, "could not create a service")
	}

	return service.Run(context.Background())
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
