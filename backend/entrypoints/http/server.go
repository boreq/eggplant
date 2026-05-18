package http

import (
	"context"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

type Server struct {
	handler http.Handler
	log     logging.Logger
}

func NewServer(handler http.Handler) *Server {
	return &Server{
		handler: handler,
		log:     logging.New("ports/http.Server"),
	}
}

func (s *Server) Serve(ctx context.Context, address string) error {
	handler := applyCORSMiddleware(s.handler)

	// Add GZIP middleware
	handler = gziphandler.GzipHandler(handler)

	httpServer := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		_ = httpServer.Close()
	}()

	s.log.Info("starting listening", "address", address)
	if err := httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
	}

	return nil
}
