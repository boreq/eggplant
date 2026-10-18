package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/boreq/eggplant/internal/logging"
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
	gzipHandler := gziphandler.GzipHandler(s.handler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			s.handler.ServeHTTP(w, r)
			return
		}
		gzipHandler.ServeHTTP(w, r)
	})

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
		return errors.Wrap(err, "error starting listening")
	}

	return nil
}
