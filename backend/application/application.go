package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/application/remote"
)

type Application struct {
	auth    *auth.Auth
	music   *music.Music
	queries *Queries
	remote  *remote.Remote
}

func NewApplication(
	auth *auth.Auth,
	music *music.Music,
	queries *Queries,
	remote *remote.Remote,
) *Application {
	return &Application{
		auth:    auth,
		music:   music,
		queries: queries,
		remote:  remote,
	}
}

func (a Application) Auth() *auth.Auth {
	return a.auth
}

func (a Application) Music() *music.Music {
	return a.music
}

func (a Application) Queries() *Queries {
	return a.queries
}

func (a Application) Remote() *remote.Remote {
	return a.remote
}

type Queries struct {
	stats   *queries.StatsHandler
	version *queries.VersionHandler
}

func NewQueries(
	stats *queries.StatsHandler,
	version *queries.VersionHandler,
) *Queries {
	return &Queries{
		stats:   stats,
		version: version,
	}
}

func (q Queries) Stats() *queries.StatsHandler {
	return q.stats
}

func (q Queries) Version() *queries.VersionHandler {
	return q.version
}
