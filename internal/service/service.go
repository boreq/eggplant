package service

import (
	"context"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	"github.com/boreq/eggplant/internal/config"
	httpPort "github.com/boreq/eggplant/ports/http"
	"github.com/pkg/errors"
)

const updateLastSeenEvery = 5 * time.Minute

type Service struct {
	httpServer      *httpPort.Server
	lastSeenUpdater *auth.LastSeenUpdater
	conf            *config.Config
}

func NewService(
	httpServer *httpPort.Server,
	lastSeenUpdater *auth.LastSeenUpdater,
	conf *config.Config,
) *Service {
	return &Service{
		httpServer:      httpServer,
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

	for i := 0; i < 2; i++ {
		if err := <-ch; err != nil {
			return errors.Wrap(err, "error during shutdown")
		}
	}

	return nil
}
