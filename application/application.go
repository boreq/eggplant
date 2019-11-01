package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/queries"
)

type Application struct {
	Auth     Auth
	Commands Commands
	Queries  Queries
}

type Auth struct {
	RegisterInitial  *auth.RegisterInitialHandler
	Register         *auth.RegisterHandler
	Login            *auth.LoginHandler
	Logout           *auth.LogoutHandler
	CheckAccessToken *auth.CheckAccessTokenHandler
	List             *auth.ListHandler
	CreateInvitation *auth.CreateInvitationHandler
}

type Commands struct {
}

type Queries struct {
	Stats *queries.StatsHandler
}
