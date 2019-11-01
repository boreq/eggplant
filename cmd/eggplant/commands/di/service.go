package di

import (
	httpPort "github.com/boreq/eggplant/pkg/service/ports/http"
)

type Service struct {
	HTTPServer *httpPort.Server
}

func NewService(httpServer *httpPort.Server) *Service {
	return &Service{
		HTTPServer: httpServer,
	}
}
