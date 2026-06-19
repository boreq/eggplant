package service

import (
	"context"
	"time"

	authApp "github.com/boreq/eggplant/application/auth"
	remoteApp "github.com/boreq/eggplant/application/remote"
	fsEntrypoint "github.com/boreq/eggplant/entrypoints/filesystem"
	httpEntrypoint "github.com/boreq/eggplant/entrypoints/http"
	outboxEntrypoint "github.com/boreq/eggplant/entrypoints/outbox"
	timersEntrypoint "github.com/boreq/eggplant/entrypoints/timers"
	"github.com/boreq/eggplant/internal/config"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	updateLastSeenEvery = 5 * time.Minute
	checkRemotesEvery   = 15 * time.Minute
)

type Service struct {
	httpServer       *httpEntrypoint.Server
	fsListener       *fsEntrypoint.Listener
	lastSeenTimer    *timersEntrypoint.LastSeen
	healthcheckTimer *timersEntrypoint.Healthcheck
	outboxListener   *outboxEntrypoint.Listener
	conf             *config.Config
}

func NewService(
	httpServer *httpEntrypoint.Server,
	fsListener *fsEntrypoint.Listener,
	persistLastSeen *authApp.PersistLastSeenHandler,
	checkRemotes *remoteApp.CheckRemotesHandler,
	outboxListener *outboxEntrypoint.Listener,
	conf *config.Config,
) *Service {
	return &Service{
		httpServer:       httpServer,
		fsListener:       fsListener,
		lastSeenTimer:    timersEntrypoint.NewLastSeen(persistLastSeen, updateLastSeenEvery),
		healthcheckTimer: timersEntrypoint.NewHealthcheck(checkRemotes, checkRemotesEvery),
		outboxListener:   outboxListener,
		conf:             conf,
	}
}

func (s *Service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.httpServer.Serve(ctx, s.conf.ServeAddress)
	})

	g.Go(func() error {
		return s.lastSeenTimer.Run(ctx)
	})

	g.Go(func() error {
		return s.healthcheckTimer.Run(ctx)
	})

	g.Go(func() error {
		return s.fsListener.Run(ctx)
	})

	g.Go(func() error {
		return s.outboxListener.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "error during shutdown")
	}

	return nil
}
