package commands

import (
	"github.com/boreq/eggplant/config"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/server"
	"github.com/boreq/guinea"
	"github.com/pkg/errors"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to a directory containing your music",
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

	errC := make(chan error)

	lib, err := library.Open(c.Arguments[0])
	if err != nil {
		return errors.Wrap(err, "opening library failed")
	}

	go func() {
		errC <- server.Serve(lib, conf.ServeAddress)
	}()

	return <-errC
}
