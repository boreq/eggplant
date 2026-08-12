package queries

import (
	"github.com/boreq/eggplant/application/accessctx"
)

type Version struct {
	version string
}

func (v Version) String() string {
	return v.version
}

type VersionHandler struct {
	version string
}

func NewVersionHandler(version string) *VersionHandler {
	return &VersionHandler{version: version}
}

func (h *VersionHandler) Execute(accessCtx accessctx.AccessContext) (Version, error) {
	if !accessCtx.Can(accessctx.PermissionSeeVersion) {
		return Version{}, accessctx.ErrPermissionDenied
	}
	return Version{version: h.version}, nil
}
