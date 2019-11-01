package wire

import (
	"github.com/boreq/eggplant/cmd/eggplant/commands/config"
	"github.com/boreq/eggplant/errors"
	"github.com/boreq/eggplant/pkg/service/adapters/music/library"
	"github.com/boreq/eggplant/pkg/service/adapters/music/scanner"
	"github.com/boreq/eggplant/pkg/service/adapters/music/store"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var musicSet = wire.NewSet(
	newLibrary,
	newTrackStore,
	newThumbnailStore,
)

func newLibrary(trackStore *store.TrackStore, thumbnailStore *store.Store, conf *config.Config) (*library.Library, error) {
	scan, err := scanner.New(conf.MusicDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a scanner")
	}

	ch, err := scan.Start()
	if err != nil {
		return nil, errors.Wrap(err, "could not start a scanner")
	}

	lib, err := library.New(ch, thumbnailStore, trackStore)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a library")
	}

	return lib, nil
}

func newTrackStore(conf *config.Config) (*store.TrackStore, error) {
	trackStore, err := store.NewTrackStore(conf.DataDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a track store")
	}
	return trackStore, nil
}

func newThumbnailStore(conf *config.Config) (*store.Store, error) {
	thumbnailStore, err := store.NewThumbnailStore(conf.DataDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a thumbnail store")
	}
	return thumbnailStore, nil
}
