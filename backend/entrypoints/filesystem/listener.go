package filesystem

import (
	"context"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/logging"
)

type Listener struct {
	processUpdate *music.ProcessUpdateHandler
	updates       <-chan scanner.Album
	log           logging.Logger
}

func NewListener(processUpdate *music.ProcessUpdateHandler, updates <-chan scanner.Album) *Listener {
	return &Listener{
		processUpdate: processUpdate,
		updates:       updates,
		log:           logging.New("entrypoints/filesystem"),
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
			if err := l.processUpdate.Execute(scan); err != nil {
				l.log.Error("could not process a scanner update", "err", err)
			}
		}
	}
}
