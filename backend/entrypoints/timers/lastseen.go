package timers

import (
	"context"
	"time"

	"github.com/boreq/eggplant/internal/logging"
)

type PersistLastSeenHandler interface {
	Execute(ctx context.Context) error
}

type LastSeen struct {
	persistLastSeen PersistLastSeenHandler
	interval        time.Duration
	log             logging.Logger
}

func NewLastSeen(persistLastSeen PersistLastSeenHandler, interval time.Duration) *LastSeen {
	return &LastSeen{
		persistLastSeen: persistLastSeen,
		interval:        interval,
		log:             logging.New("entrypoints/timers.LastSeen"),
	}
}

func (t *LastSeen) Run(ctx context.Context) error {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := t.persistLastSeen.Execute(ctx); err != nil {
				t.log.Error("last seen persist failed", "err", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
