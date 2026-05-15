//go:build insecurecors

package http

import (
	"net/http"

	"github.com/rs/cors"
)

func applyCORSMiddleware(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowOriginFunc:  func(string) bool { return true },
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(h)
}
