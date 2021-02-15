package wire

import (
	"github.com/boreq/eggplant/adapters/music/library"
	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/errors"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var musicSet = wire.NewSet(
	newLibrary,
	newTrackStore,
	newThumbnailStore,
	newScannerConfig,
	library.NewDelimiterAccessLoader,
	library.NewIdGenerator,

	wire.Bind(new(library.AccessLoader), new(*library.DelimiterAccessLoader)),
	wire.Bind(new(library.TrackStore), new(*store.TrackStore)),
	wire.Bind(new(library.ThumbnailStore), new(*store.Store)),
	wire.Bind(new(music.TrackStore), new(*store.TrackStore)),
	wire.Bind(new(music.ThumbnailStore), new(*store.Store)),
	wire.Bind(new(music.Library), new(*library.Library)),
	wire.Bind(new(queries.TrackStore), new(*store.TrackStore)),
	wire.Bind(new(queries.ThumbnailStore), new(*store.Store)),
)

func newLibrary(
	accessLoader library.AccessLoader,
	trackStore library.TrackStore,
	thumbnailStore library.ThumbnailStore,
	idGenerator library.IdGenerator,
	conf *config.Config,
	scannerConf scanner.Config,
) (*library.Library, error) {
	scan, err := scanner.New(conf.MusicDirectory, scannerConf)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a scanner")
	}

	ch, err := scan.Start()
	if err != nil {
		return nil, errors.Wrap(err, "could not start a scanner")
	}

	lib, err := library.New(ch, trackStore, thumbnailStore, accessLoader, idGenerator)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a library")
	}

	return lib, nil
}

func newTrackStore(conf *config.Config) (*store.TrackStore, error) {
	trackStore, err := store.NewTrackStore(conf.CacheDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a track store")
	}
	return trackStore, nil
}

func newThumbnailStore(conf *config.Config) (*store.Store, error) {
	thumbnailStore, err := store.NewThumbnailStore(conf.CacheDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a thumbnail store")
	}
	return thumbnailStore, nil
}

func newScannerConfig(conf *config.Config) scanner.Config {
	return scanner.Config{
		TrackExtensions:     conf.TrackExtensions,
		ThumbnailStems:      conf.ThumbnailStems,
		ThumbnailExtensions: conf.ThumbnailExtensions,
	}
}
