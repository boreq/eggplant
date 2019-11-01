package http

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/boreq/eggplant/logging"
	"github.com/rs/cors"
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

func (s *Server) Serve(address string) error {
	// Add CORS middleware
	handler := cors.AllowAll().Handler(s.handler)

	// Add GZIP middleware
	handler = gziphandler.GzipHandler(handler)

	s.log.Info("starting listening", "address", address)
	return http.ListenAndServe(address, handler)
}
