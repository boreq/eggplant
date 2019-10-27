package commands

import (
	"github.com/boreq/eggplant/config"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/scanner"
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
	conf.MusicDirectory = c.Arguments[0]
	conf.CacheDirectory = c.Arguments[1]

	errC := make(chan error)

	scan, err := scanner.New(conf.MusicDirectory)
	if err != nil {
		return errors.Wrap(err, "could not create a scanner")
	}

	ch, err := scan.Start()
	if err != nil {
		return errors.Wrap(err, "could not start a scanner")
	}

	trackStore, err := store.NewTrackStore(conf.CacheDirectory)
	if err != nil {
		return errors.Wrap(err, "could not create a track store")
	}

	thumbnailStore, err := store.NewThumbnailStore(conf.CacheDirectory)
	if err != nil {
		return errors.Wrap(err, "could not create a thumbnail store")
	}

	lib, err := library.New(ch, thumbnailStore, trackStore)
	if err != nil {
		return errors.Wrap(err, "could not create a library")
	}

	go func() {
		errC <- server.Serve(lib, trackStore, thumbnailStore, conf.ServeAddress)
	}()

	return <-errC
}
