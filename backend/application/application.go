package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
)

type Application struct {
	Auth    auth.Auth
	Music   Music
	Queries Queries
}

type Music struct {
	Thumbnail *music.ThumbnailHandler
	Track     *music.TrackHandler
	Browse    *music.BrowseHandler
	Search    *music.SearchHandler
}

type Queries struct {
	Stats *queries.StatsHandler
}
