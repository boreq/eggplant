package queries

import (
	"github.com/boreq/eggplant/application/accessctx"
)

type Version struct {
	backend  string
	frontend string
}

func NewVersion(backend, frontend string) Version {
	return Version{
		backend:  backend,
		frontend: frontend,
	}
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
	return NewVersion(h.backend, h.frontend), nil
}
