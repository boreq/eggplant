package fixture

import (
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music/library"
)

type adminAccessContext struct{}

func (adminAccessContext) Can(accessctx.Permission) bool {
	return true
}

func (adminAccessContext) CanSee(library.Visibility) bool {
	return true
}

func AdminAccessContext() accessctx.AccessContext {
	return adminAccessContext{}
}
