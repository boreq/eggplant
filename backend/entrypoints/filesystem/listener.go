package filesystem

import (
	"context"
	"time"

	"github.com/boreq/eggplant/internal/logging"
)

const retryDelay = 30 * time.Second

type LoadLibraryHandler interface {
	Execute(ctx context.Context) error
}

type Listener struct {
	loadLibrary LoadLibraryHandler
	updates     <-chan struct{}
	log         logging.Logger
}

func NewListener(loadLibrary LoadLibraryHandler, updates <-chan struct{}) *Listener {
	return &Listener{
		loadLibrary: loadLibrary,
		updates:     updates,
		log:         logging.New("entrypoints/filesystem"),
	}
}

func (l *Listener) Run(ctx context.Context) error {
	dirty := false

	for {
		if !dirty {
			select {
			case <-ctx.Done():
				return nil
			case _, ok := <-l.updates:
				if !ok {
					return nil
				}
				dirty = true
			}
			continue
		}

		if err := l.loadLibrary.Execute(ctx); err != nil {
			l.log.Error("could not load the library, will retry", "err", err, "delay", retryDelay)
			if !l.waitOrDrain(ctx, retryDelay) {
				return nil
			}
			continue
		}

		dirty = false
	}
}

func (l *Listener) waitOrDrain(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case _, ok := <-l.updates:
			if !ok {
				return false
			}
		case <-timer.C:
			return true
		}
	}
}
