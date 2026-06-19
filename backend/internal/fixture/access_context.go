package fixture

import "github.com/boreq/eggplant/application/accessctx"

type adminAccessContext struct{}

func (adminAccessContext) Can(accessctx.Permission) bool {
	return true
}

func AdminAccessContext() accessctx.AccessContext {
	return adminAccessContext{}
}
