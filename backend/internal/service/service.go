package service

import (
	"context"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	fsEntrypoint "github.com/boreq/eggplant/entrypoints/filesystem"
	httpEntrypoint "github.com/boreq/eggplant/entrypoints/http"
	"github.com/boreq/eggplant/internal/config"
	"github.com/pkg/errors"
)

const updateLastSeenEvery = 5 * time.Minute

type Service struct {
	httpServer      *httpEntrypoint.Server
	fsListener      *fsEntrypoint.Listener
	lastSeenUpdater *auth.LastSeenUpdater
	conf            *config.Config
}

func NewService(
	httpServer *httpEntrypoint.Server,
	fsListener *fsEntrypoint.Listener,
	lastSeenUpdater *auth.LastSeenUpdater,
	conf *config.Config,
) *Service {
	return &Service{
		httpServer:      httpServer,
		fsListener:      fsListener,
		lastSeenUpdater: lastSeenUpdater,
		conf:            conf,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ch := make(chan error)

	go func() {
		ch <- s.httpServer.Serve(ctx, s.conf.ServeAddress)
	}()

	go func() {
		s.lastSeenUpdater.Run(ctx, updateLastSeenEvery)
		ch <- nil
	}()

	go func() {
		ch <- s.fsListener.Run(ctx)
	}()

	for i := 0; i < 3; i++ {
		if err := <-ch; err != nil {
			return errors.Wrap(err, "error during shutdown")
		}
	}

	return nil
}
