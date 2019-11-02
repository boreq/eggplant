package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
)

type Application struct {
	Auth    Auth
	Music   Music
	Queries Queries
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

type Music struct {
	Thumbnail *music.ThumbnailHandler
	Track     *music.TrackHandler
	Browse    *music.BrowseHandler
}

type Queries struct {
	Stats *queries.StatsHandler
}
