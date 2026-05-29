package wire

import (
	"context"

	"github.com/boreq/eggplant/adapters/music/library"
	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/adapters/music/tracks"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/errors"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var musicSet = wire.NewSet(
	newScanner,
	newDirectoryWatcherUpdates,
	newTrackStore,
	newThumbnailStore,
	newScannerConfig,
	library.NewDelimiterAccessLoader,
	library.NewInMemoryRepository,
	tracks.NewFFProbe,
	tracks.NewDurationStore,

	wire.Bind(new(music.Scanner), new(*scanner.Scanner)),
	wire.Bind(new(music.AccessLoader), new(*library.DelimiterAccessLoader)),
	wire.Bind(new(music.TrackConverter), new(*tracks.Converter)),
	wire.Bind(new(music.ThumbnailStore), new(*store.ThumbnailStore)),
	wire.Bind(new(tracks.DurationChecker), new(*tracks.FFProbe)),
	wire.Bind(new(music.TrackDurationStore), new(*tracks.DurationStore)),
	wire.Bind(new(music.LibraryRepository), new(*library.InMemoryRepository)),
	wire.Bind(new(queries.TrackStore), new(*tracks.Converter)),
	wire.Bind(new(queries.ThumbnailStore), new(*store.ThumbnailStore)),
)

func newScanner(conf *config.Config, scannerConf scanner.Config) (*scanner.Scanner, error) {
	s, err := scanner.New(conf.MusicDirectory, scannerConf)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a scanner")
	}
	return s, nil
}

func newDirectoryWatcherUpdates(conf *config.Config) (<-chan struct{}, error) {
	w := scanner.NewDirectoryWatcher(conf.MusicDirectory)
	ch, err := w.Start()
	if err != nil {
		return nil, errors.Wrap(err, "could not start the directory watcher")
	}
	return ch, nil
}

func newTrackStore(ctx context.Context, conf *config.Config) (*tracks.Converter, error) {
	converter, err := tracks.NewConverter(ctx, conf.CacheDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a track converter")
	}
	return converter, nil
}

func newThumbnailStore(ctx context.Context, conf *config.Config) (*store.ThumbnailStore, error) {
	thumbnailStore, err := store.NewThumbnailStore(ctx, conf.CacheDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a thumbnail store")
	}
	return thumbnailStore, nil
}

func newScannerConfig(conf *config.Config) (scanner.Config, error) {
	trackExts, err := toFileExtensions(conf.TrackExtensions)
	if err != nil {
		return scanner.Config{}, errors.Wrap(err, "could not parse track extensions")
	}

	thumbnailExts, err := toFileExtensions(conf.ThumbnailExtensions)
	if err != nil {
		return scanner.Config{}, errors.Wrap(err, "could not parse thumbnail extensions")
	}

	thumbnailStems, err := toThumbnailStems(conf.ThumbnailStems)
	if err != nil {
		return scanner.Config{}, errors.Wrap(err, "could not parse thumbnail stems")
	}

	return scanner.NewConfig(trackExts, thumbnailStems, thumbnailExts)
}

func toFileExtensions(values []string) ([]musicdomain.FileExtension, error) {
	out := make([]musicdomain.FileExtension, 0, len(values))
	for _, v := range values {
		ext, err := musicdomain.NewFileExtension(v)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid extension '%s'", v)
		}
		out = append(out, ext)
	}
	return out, nil
}

func toThumbnailStems(values []string) ([]scanner.ThumbnailStem, error) {
	out := make([]scanner.ThumbnailStem, 0, len(values))
	for _, v := range values {
		stem, err := scanner.NewThumbnailStem(v)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid stem '%s'", v)
		}
		out = append(out, stem)
	}
	return out, nil
}
