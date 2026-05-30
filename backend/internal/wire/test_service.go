package wire

import (
	"net/http"

	"github.com/boreq/eggplant/application"
	bolt "go.etcd.io/bbolt"
)

type TestHTTPService struct {
	Handler http.Handler
	App     *application.Application
	DB      *bolt.DB
}

func newTestHTTPService(handler http.Handler, app *application.Application, db *bolt.DB) *TestHTTPService {
	return &TestHTTPService{
		Handler: handler,
		App:     app,
		DB:      db,
	}
}
