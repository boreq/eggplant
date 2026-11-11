package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/application/remotecontrol"
)

type Application struct {
	auth    *auth.Auth
	control *remotecontrol.Control
	music   *music.Music
	queries *Queries
	remote  *remote.Remote
}

func NewApplication(
	auth *auth.Auth,
	control *remotecontrol.Control,
	music *music.Music,
	queries *Queries,
	remote *remote.Remote,
) *Application {
	return &Application{
		auth:    auth,
		control: control,
		music:   music,
		queries: queries,
		remote:  remote,
	}
}

func (a Application) Auth() *auth.Auth {
	return a.auth
}

func (a Application) Control() *remotecontrol.Control {
	return a.control
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
