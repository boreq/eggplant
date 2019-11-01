package application

import (
	"github.com/boreq/eggplant/pkg/service/application/auth"
	"github.com/boreq/eggplant/pkg/service/application/queries"
)

type Application struct {
	Auth     Auth
	Commands Commands
	Queries  Queries
}

type Auth struct {
	RegisterInitial  *auth.RegisterInitialHandler
	Login            *auth.LoginHandler
	Logout           *auth.LogoutHandler
	CheckAccessToken *auth.CheckAccessTokenHandler
	List             *auth.ListHandler
}

type Commands struct {
}

type Queries struct {
	Stats *queries.StatsHandler
}
