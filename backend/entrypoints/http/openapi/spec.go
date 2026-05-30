package openapi

import _ "embed"

// Spec is the raw OpenAPI 3 specification in YAML. It is the source of truth
// for the HTTP API: it is served as the API reference and is used to generate
// the Go types and HTTP client in this package.
//
//go:embed openapi.yaml
var Spec []byte
