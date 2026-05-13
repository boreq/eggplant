//go:build !insecurecors

package http

import "net/http"

func applyCORSMiddleware(h http.Handler) http.Handler {
	return h
}
