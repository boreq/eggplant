package service

import (
	httpPort "github.com/boreq/eggplant/ports/http"
)

type Service struct {
	HTTPServer *httpPort.Server
}

func NewService(httpServer *httpPort.Server) *Service {
	return &Service{
		HTTPServer: httpServer,
	}
}
