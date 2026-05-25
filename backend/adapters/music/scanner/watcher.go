package scanner

import (
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	"github.com/radovskyb/watcher"
)

type DirectoryWatcher struct {
	directory string
	logger    logging.Logger
}

func NewDirectoryWatcher(directory string) *DirectoryWatcher {
	return &DirectoryWatcher{
		directory: directory,
		logger:    logging.New("directory_watcher"),
	}
}

// Start begins watching and returns a channel that receives a signal on every
// detected change, plus an initial signal upon start.
func (d *DirectoryWatcher) Start() (<-chan struct{}, error) {
	w := watcher.New()
	w.SetMaxEvents(1)

	if err := w.AddRecursive(d.directory); err != nil {
		return nil, errors.Wrap(err, "could not add a recursive watcher")
	}

	go func() {
		if err := w.Start(time.Second * 10); err != nil {
			d.logger.Error("error starting the watcher", "err", err)
		}
	}()

	ch := make(chan struct{}, 1)
	go func() {
		defer close(ch)
		ch <- struct{}{}

		for {
			select {
			case _, ok := <-w.Event:
				if !ok {
					return
				}
				ch <- struct{}{}
			case err := <-w.Error:
				d.logger.Error("watcher error", "err", err)
			case <-w.Closed:
				return
			}
		}
	}()
	return ch, nil
}
