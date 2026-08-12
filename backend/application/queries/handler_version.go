package queries

import (
	"github.com/boreq/eggplant/application/accessctx"
)

type Version struct {
	backend  string
	frontend string
}

func (v Version) Backend() string {
	return v.backend
}

func (v Version) Frontend() string {
	return v.frontend
}

type VersionHandler struct {
	backend  string
	frontend string
}

func NewVersionHandler(backend, frontend string) *VersionHandler {
	return &VersionHandler{backend: backend, frontend: frontend}
}

func (h *VersionHandler) Execute(accessCtx accessctx.AccessContext) (Version, error) {
	if !accessCtx.Can(accessctx.PermissionSeeVersion) {
		return Version{}, accessctx.ErrPermissionDenied
	}
	return Version{backend: h.backend, frontend: h.frontend}, nil
}
