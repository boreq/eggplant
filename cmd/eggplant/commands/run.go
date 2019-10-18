package commands

import (
	"github.com/boreq/eggplant/config"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/loader"
	"github.com/boreq/eggplant/server"
	"github.com/boreq/eggplant/store"
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
		{
			Name:        "cache_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to a directory which will be used for caching",
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

	loader, err := loader.New(c.Arguments[0])
	if err != nil {
		return errors.Wrap(err, "could not create a loader")
	}

	ch, err := loader.Start()
	if err != nil {
		return errors.Wrap(err, "could not start a loader")
	}

	trackStore, err := store.NewTrackStore(c.Arguments[1])
	if err != nil {
		return errors.Wrap(err, "creating store failed")
	}

	thumbnailStore, err := store.NewThumbnailStore(c.Arguments[1])
	if err != nil {
		return errors.Wrap(err, "creating thumbnail store failed")
	}

	lib, err := library.New(ch, thumbnailStore, trackStore)
	if err != nil {
		return errors.Wrap(err, "opening library failed")
	}

	go func() {
		errC <- server.Serve(lib, trackStore, thumbnailStore, conf.ServeAddress)
	}()

	return <-errC
}
