package filesystem

import (
	"context"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain/scanner"
	"github.com/boreq/eggplant/logging"
)

type Listener struct {
	buildLibrary *music.BuildLibraryHandler
	updates      <-chan scanner.FoundRootAlbum
	log          logging.Logger
}

func NewListener(buildLibrary *music.BuildLibraryHandler, updates <-chan scanner.FoundRootAlbum) *Listener {
	return &Listener{
		buildLibrary: buildLibrary,
		updates:      updates,
		log:          logging.New("entrypoints/filesystem"),
	}
}

func (l *Listener) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case scan, ok := <-l.updates:
			if !ok {
				return nil
			}
			if err := l.buildLibrary.Execute(scan); err != nil {
				l.log.Error("could not process a scanner update", "err", err)
			}
		}
	}
}
