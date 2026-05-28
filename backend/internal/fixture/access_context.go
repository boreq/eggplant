package fixture

import "github.com/boreq/eggplant/application/auth"

type adminAccessContext struct{}

func (adminAccessContext) Can(auth.Permission) bool {
	return true
}

func AdminAccessContext() auth.AccessContext {
	return adminAccessContext{}
}
