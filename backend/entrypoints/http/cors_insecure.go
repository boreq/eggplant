//go:build insecurecors

package http

import (
	"net/http"

	"github.com/rs/cors"
)

func applyCORSMiddleware(h http.Handler) http.Handler {
	return cors.AllowAll().Handler(h)
}
