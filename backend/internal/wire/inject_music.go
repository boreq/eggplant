package wire

import (
	"context"

	library2 "github.com/boreq/eggplant/adapters/music/library"
	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/music/library"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/errors"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var musicSet = wire.NewSet(
	newLibrary,
	newScannerUpdates,
	newTrackStore,
	newThumbnailStore,
	newScannerConfig,
	library2.NewDelimiterAccessLoader,
	library2.NewIdGenerator,

	wire.Bind(new(library.AccessLoader), new(*library2.DelimiterAccessLoader)),
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
) (*library.Library, error) {
	lib, err := library.New(trackStore, thumbnailStore, accessLoader, idGenerator)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a library")
	}
	return lib, nil
}

func newScannerUpdates(conf *config.Config, scannerConf scanner.Config) (<-chan scanner.Album, error) {
	scan, err := scanner.New(conf.MusicDirectory, scannerConf)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a scanner")
	}
	ch, err := scan.Start()
	if err != nil {
		return nil, errors.Wrap(err, "could not start a scanner")
	}
	return ch, nil
}

func newTrackStore(ctx context.Context, conf *config.Config) (*store.TrackStore, error) {
	trackStore, err := store.NewTrackStore(ctx, conf.CacheDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a track store")
	}
	return trackStore, nil
}

func newThumbnailStore(ctx context.Context, conf *config.Config) (*store.Store, error) {
	thumbnailStore, err := store.NewThumbnailStore(ctx, conf.CacheDirectory)
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
