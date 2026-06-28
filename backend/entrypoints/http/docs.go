package http

import (
	"net/http"

	"github.com/boreq/eggplant/adapters/openapi"
)

// redocPage renders the API reference using Redoc. The Redoc bundle is loaded
// from a CDN; the spec itself is served locally from /api/openapi.yaml.
const redocPage = `<!DOCTYPE html>
<html>
  <head>
    <title>Eggplant API</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <style>body { margin: 0; padding: 0; }</style>
  </head>
  <body>
    <redoc spec-url="/api/openapi.yaml"></redoc>
    <script src="https://cdn.redocly.com/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>`

func (h *Handler) openapiSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(openapi.Spec)
}

func (h *Handler) docs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(redocPage))
}
