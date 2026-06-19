package wire

import (
	"net/http"

	"github.com/boreq/eggplant/application"
	outboxEntrypoint "github.com/boreq/eggplant/entrypoints/outbox"
	bolt "go.etcd.io/bbolt"
)

type TestHTTPService struct {
	Handler        http.Handler
	App            *application.Application
	OutboxListener *outboxEntrypoint.Listener
	DB             *bolt.DB
}

func newTestHTTPService(handler http.Handler, app *application.Application, outboxListener *outboxEntrypoint.Listener, db *bolt.DB) *TestHTTPService {
	return &TestHTTPService{
		Handler:        handler,
		App:            app,
		OutboxListener: outboxListener,
		DB:             db,
	}
}
