package commands

import (
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/boreq/errors"
	"github.com/boreq/guinea"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "music_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to a directory containing your music",
		},
		{
			Name:        "data_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to a directory which will be used for data storage",
		},
	},
	Options: []guinea.Option{
		guinea.Option{
			Name:        "address",
			Type:        guinea.String,
			Description: "Serve address",
			Default:     config.Default().ServeAddress,
		},
	},
	ShortDescription: "serves your music",
}

func runRun(c guinea.Context) error {
	conf := config.Default()
	conf.ServeAddress = c.Options["address"].Str()
	conf.MusicDirectory = c.Arguments[0]
	conf.DataDirectory = c.Arguments[1]

	service, err := wire.BuildService(conf)
	if err != nil {
		return errors.Wrap(err, "could not create a service")
	}

	return service.HTTPServer.Serve(conf.ServeAddress)
}
