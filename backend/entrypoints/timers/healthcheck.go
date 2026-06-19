package timers

import (
	"context"
	"time"

	"github.com/boreq/eggplant/internal/logging"
)

type CheckRemotesHandler interface {
	Execute(ctx context.Context) error
}

type Healthcheck struct {
	checkRemotes CheckRemotesHandler
	interval     time.Duration
	log          logging.Logger
}

func NewHealthcheck(checkRemotes CheckRemotesHandler, interval time.Duration) *Healthcheck {
	return &Healthcheck{
		checkRemotes: checkRemotes,
		interval:     interval,
		log:          logging.New("entrypoints/timers.Healthcheck"),
	}
}

func (t *Healthcheck) Run(ctx context.Context) error {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		if err := t.checkRemotes.Execute(ctx); err != nil {
			t.log.Error("healthcheck run failed", "err", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil
		}
	}
}
