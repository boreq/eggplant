package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/application/remote"
)

type Application struct {
	Auth    auth.Auth
	Music   music.Music
	Queries Queries
	Remote  remote.Remote
}

type Queries struct {
	Stats   *queries.StatsHandler
	Version *queries.VersionHandler
}
