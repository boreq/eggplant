package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/application/remotecontrol"
)

type Application struct {
	Auth    auth.Auth
	Control remotecontrol.Control
	Music   music.Music
	Queries Queries
	Remote  remote.Remote
}

type Queries struct {
	Stats   *queries.StatsHandler
	Version *queries.VersionHandler
}
